// internal/forwarder/manager.go
package forwarder

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"tcp-proxy-bridge/internal/config"
	"tcp-proxy-bridge/internal/database"
	"tcp-proxy-bridge/internal/metrics"
)

// Manager 转发器管理器
// 负责管理多个目标服务器的消息转发工作
type Manager struct {
	config  *config.ForwarderConfig  // 转发器配置
	db      *database.Postgres       // 数据库实例
	targets []*database.TargetServer // 目标服务器列表

	workers   map[string]*Worker // 工作器映射表
	mu        sync.RWMutex       // 读写锁
	isRunning bool               // 运行状态

	shutdownChan chan struct{}  // 关闭信号通道
	wg           sync.WaitGroup // 等待组
}

// Worker 转发工作器
// 负责向特定目标服务器转发消息
type Worker struct {
	target       *database.TargetServer  // 目标服务器
	db           *database.Postgres      // 数据库实例
	config       *config.ForwarderConfig // 转发配置
	isRunning    bool                    // 运行状态
	shutdownChan chan struct{}           // 关闭信号通道
	wg           sync.WaitGroup          // 等待组
}

// NewManager 创建转发器管理器
// 参数: cfg - 转发器配置, db - 数据库实例, targets - 目标服务器列表
// 返回: 转发器管理器实例
func NewManager(cfg *config.ForwarderConfig, db *database.Postgres, targets []*database.TargetServer) *Manager {
	return &Manager{
		config:       cfg,
		db:           db,
		targets:      targets,
		workers:      make(map[string]*Worker),
		shutdownChan: make(chan struct{}),
	}
}

// Start 启动转发器管理器
// 参数: ctx - 上下文
// 返回: 错误信息
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("forwarder manager is already running")
	}

	log.Printf("Starting forwarder manager with %d target servers", len(m.targets))

	// 为每个启用的目标服务器创建工作器
	for _, target := range m.targets {
		if target.Enabled {
			worker := NewWorker(target, m.db, m.config)
			m.workers[target.ID] = worker
			go worker.Start(ctx)
			log.Printf("Started worker for target server: %s (%s)", target.Name, target.Address)
		}
	}

	m.isRunning = true
	log.Printf("Forwarder manager started successfully with %d workers", len(m.workers))

	return nil
}

// Stop 停止转发器管理器
// 参数: ctx - 上下文
// 返回: 错误信息
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return nil
	}

	log.Println("Stopping forwarder manager...")

	// 发送关闭信号
	close(m.shutdownChan)

	// 停止所有工作器
	var stopWg sync.WaitGroup
	for targetID, worker := range m.workers {
		stopWg.Add(1)
		go func(w *Worker, id string) {
			defer stopWg.Done()
			w.Stop(ctx)
			log.Printf("Worker for target %s stopped", id)
		}(worker, targetID)
	}

	// 等待所有工作器停止
	stopWg.Wait()
	m.isRunning = false

	log.Println("Forwarder manager stopped successfully")
	return nil
}

// NewWorker 创建工作器实例
// 参数: target - 目标服务器, db - 数据库实例, cfg - 转发配置
// 返回: 工作器实例
func NewWorker(target *database.TargetServer, db *database.Postgres, cfg *config.ForwarderConfig) *Worker {
	return &Worker{
		target:       target,
		db:           db,
		config:       cfg,
		shutdownChan: make(chan struct{}),
	}
}

// Start 启动工作器
// 参数: ctx - 上下文
func (w *Worker) Start(ctx context.Context) {
	w.isRunning = true
	w.wg.Add(1)

	log.Printf("Starting message forwarding worker for target: %s", w.target.Name)

	// 启动消息处理循环
	go w.processMessages(ctx)
}

// Stop 停止工作器
// 参数: ctx - 上下文
func (w *Worker) Stop(ctx context.Context) {
	if !w.isRunning {
		return
	}

	log.Printf("Stopping worker for target: %s", w.target.Name)

	// 发送关闭信号
	close(w.shutdownChan)

	// 等待处理循环结束
	w.wg.Wait()
	w.isRunning = false

	log.Printf("Worker for target %s stopped", w.target.Name)
}

// processMessages 处理消息的主循环
// 参数: ctx - 上下文
func (w *Worker) processMessages(ctx context.Context) {
	defer w.wg.Done()

	// 创建定时器，定期处理消息
	ticker := time.NewTicker(w.config.ProcessInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// 上下文被取消
			log.Printf("Worker context cancelled for target: %s", w.target.Name)
			return
		case <-w.shutdownChan:
			// 收到关闭信号
			log.Printf("Worker received shutdown signal for target: %s", w.target.Name)
			return
		case <-ticker.C:
			// 定时处理一批消息
			w.processBatch(ctx)
		}
	}
}

// processBatch 处理一批消息
// 参数: ctx - 上下文
func (w *Worker) processBatch(ctx context.Context) {
	// 从数据库获取待处理的消息
	messages, err := w.db.GetPendingMessagesForTarget(w.target.ID, w.config.BatchSize)
	if err != nil {
		log.Printf("Failed to get pending messages for target %s: %v", w.target.ID, err)
		return
	}

	if len(messages) == 0 {
		// 没有待处理的消息
		return
	}

	log.Printf("Processing %d messages for target %s", len(messages), w.target.Name)

	// 使用信号量控制并发数
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, w.config.MaxProcessingWorkers)

	// 处理每条消息
	for _, msg := range messages {
		select {
		case <-ctx.Done():
			// 上下文被取消，停止处理
			return
		case <-w.shutdownChan:
			// 收到关闭信号，停止处理
			return
		default:
			wg.Add(1)
			semaphore <- struct{}{} // 获取信号量

			go func(message *database.Message) {
				defer wg.Done()
				defer func() { <-semaphore }() // 释放信号量

				w.processSingleMessage(ctx, message)
			}(msg)
		}
	}

	// 等待所有消息处理完成
	wg.Wait()

	log.Printf("Completed processing batch of %d messages for target %s", len(messages), w.target.Name)
}

// processSingleMessage 处理单条消息
// 参数: ctx - 上下文, message - 要处理的消息
func (w *Worker) processSingleMessage(ctx context.Context, message *database.Message) {
	startTime := time.Now()

	// 更新消息状态为"发送中"
	if err := w.db.UpdateDeliveryStatus(message.ID, w.target.ID, database.StatusSending, 0, nil); err != nil {
		log.Printf("Failed to update delivery status for message %d: %v", message.ID, err)
		return
	}

	// 尝试发送消息到目标服务器
	err := w.sendToTarget(message)
	processingTime := time.Since(startTime).Milliseconds()

	if err != nil {
		// 发送失败，处理重试逻辑
		w.handleSendFailure(message, err, processingTime)
	} else {
		// 发送成功，更新状态
		w.handleSendSuccess(message, processingTime)
	}
}

// sendToTarget 发送消息到目标服务器
// 参数: message - 要发送的消息
// 返回: 错误信息
func (w *Worker) sendToTarget(message *database.Message) error {
	// 建立TCP连接到目标服务器
	conn, err := net.DialTimeout("tcp", w.target.Address, w.target.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to target server %s: %v", w.target.Address, err)
	}
	defer conn.Close()

	// 设置写超时
	conn.SetWriteDeadline(time.Now().Add(w.target.Timeout))

	// 发送消息数据
	_, err = conn.Write(message.OriginalData)
	if err != nil {
		return fmt.Errorf("failed to send data to target server %s: %v", w.target.Address, err)
	}

	return nil
}

// handleSendFailure 处理发送失败的情况
// 参数: message - 消息, err - 错误信息, processingTime - 处理时间
func (w *Worker) handleSendFailure(message *database.Message, err error, processingTime int64) {
	log.Printf("Failed to send message %d to target %s: %v", message.ID, w.target.Name, err)

	// 获取当前尝试次数
	currentAttempts, getErr := w.db.GetDeliveryAttempts(message.ID, w.target.ID)
	if getErr != nil {
		log.Printf("Failed to get delivery attempts for message %d: %v", message.ID, getErr)
		return
	}

	// 计算下次重试时间
	nextRetry := w.calculateNextRetry(currentAttempts)
	errorMsg := err.Error()

	// 更新数据库状态
	updateErr := w.db.UpdateDeliveryStatus(
		message.ID,
		w.target.ID,
		database.StatusFailed,
		currentAttempts+1,
		&errorMsg,
	)

	if updateErr != nil {
		log.Printf("Failed to update failed status for message %d: %v", message.ID, updateErr)
	}

	// 更新指标：增加错误计数
	metrics.IncMessageErrors()

	log.Printf("Message %d failed after %dms, scheduled retry at %v",
		message.ID, processingTime, nextRetry)
}

// handleSendSuccess 处理发送成功的情况
// 参数: message - 消息, processingTime - 处理时间
func (w *Worker) handleSendSuccess(message *database.Message, processingTime int64) {
	// 更新数据库状态为"已发送"
	updateErr := w.db.UpdateDeliveryStatus(
		message.ID,
		w.target.ID,
		database.StatusSent,
		0, // 重置重试次数
		nil,
	)

	if updateErr != nil {
		log.Printf("Failed to update sent status for message %d: %v", message.ID, updateErr)
	} else {
		log.Printf("Successfully sent message %d to target %s in %d ms",
			message.ID, w.target.Name, processingTime)
		// 更新指标：增加成功转发计数
		metrics.IncMessagesForwarded()
	}
}

// calculateNextRetry 计算下次重试时间（指数退避算法）
// 参数: attempts - 已尝试次数
// 返回: 下次重试时间
func (w *Worker) calculateNextRetry(attempts int) time.Time {
	var delay time.Duration

	// 指数退避策略
	delay = w.config.BaseRetryInterval * time.Duration(1<<uint(attempts))

	// 限制最大重试间隔
	if delay > w.config.MaxRetryInterval {
		delay = w.config.MaxRetryInterval
	}

	return time.Now().Add(delay)
}

// GetWorkerStatus 获取工作器状态
// 返回: 工作器是否在运行
func (w *Worker) GetWorkerStatus() bool {
	return w.isRunning
}
