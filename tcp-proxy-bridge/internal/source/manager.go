// internal/source/manager.go
package source

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"tcp-proxy-bridge/internal/config"
)

// Manager 源服务器管理器
// 负责管理主备源服务器的连接和故障切换
type Manager struct {
	config        *config.SourceServers // 源服务器配置
	currentServer *config.SourceServer  // 当前使用的服务器
	mu            sync.RWMutex          // 读写锁
	isRunning     bool                  // 运行状态

	// 健康检查相关
	healthChecker *HealthChecker
	shutdownChan  chan struct{}  // 关闭信号通道
	wg            sync.WaitGroup // 等待组

	// 故障统计
	failureCounts map[string]int       // 各服务器失败次数统计
	lastFailTime  map[string]time.Time // 各服务器最后失败时间

	// 协议处理相关
	protocolHandler  *ProtocolHandler   // 协议处理器
	delimiterHandler *DelimiterHandler  // 分隔符处理器
	authManager      *AuthManager       // 身份认证管理器
	heartbeatManager *HeartbeatManager  // 心跳管理器
	dataHandler      func([]byte) error // 数据处理回调函数
}

// NewManager 创建源服务器管理器
// 参数: cfg - 完整配置
// 返回: 源服务器管理器实例
func NewManager(cfg *config.Config) *Manager {
	// 解析分隔符
	delimiterBytes, err := hex.DecodeString(cfg.Delimiter.Separator)
	if err != nil {
		log.Fatalf("Failed to decode delimiter: %v", err)
	}

	// 创建分隔符处理器
	delimiterHandler := NewDelimiterHandler(delimiterBytes, cfg.Delimiter.MaxPacketLength)

	// 创建身份认证管理器
	authManager := NewAuthManager(cfg.Authentication.Token, cfg.Authentication.SourceID, cfg.Authentication.HostID)

	// 创建心跳管理器
	heartbeatManager := NewHeartbeatManager(
		cfg.Heartbeat.Interval,
		cfg.Heartbeat.WriteIdleTimeout,
		cfg.Heartbeat.ReadIdleTimeout,
	)

	return &Manager{
		config:           &cfg.SourceServers,
		currentServer:    &cfg.SourceServers.Primary, // 默认使用主服务器
		healthChecker:    NewHealthChecker(),
		shutdownChan:     make(chan struct{}),
		failureCounts:    make(map[string]int),
		lastFailTime:     make(map[string]time.Time),
		protocolHandler:  NewProtocolHandler(cfg.Heartbeat.Interval), // 使用配置的心跳间隔
		delimiterHandler: delimiterHandler,
		authManager:      authManager,
		heartbeatManager: heartbeatManager,
	}
}

// Start 启动源服务器管理器
// 参数: ctx - 上下文
// 返回: 错误信息
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("source manager is already running")
	}

	log.Printf("Starting source server manager with primary: %s, backup: %s",
		m.config.Primary.Name, m.config.Backup.Name)

	// 启动健康检查
	m.wg.Add(1)
	go m.healthCheckLoop(ctx)

	m.isRunning = true
	log.Printf("Source server manager started successfully")
	return nil
}

// Stop 停止源服务器管理器
// 参数: ctx - 上下文
// 返回: 错误信息
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return nil
	}

	log.Println("Stopping source server manager...")

	// 发送关闭信号
	close(m.shutdownChan)

	// 等待健康检查循环结束
	m.wg.Wait()
	m.isRunning = false

	log.Println("Source server manager stopped successfully")
	return nil
}

// GetCurrentServer 获取当前使用的服务器
// 返回: 当前服务器配置
func (m *Manager) GetCurrentServer() *config.SourceServer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentServer
}

// SetDataHandler 设置数据处理回调函数
// 参数: handler - 数据处理函数
func (m *Manager) SetDataHandler(handler func([]byte) error) {
	m.dataHandler = handler
}

// ConnectToSource 连接到源服务器获取数据
// 参数: ctx - 上下文, dataHandler - 数据处理函数
// 返回: 错误信息
func (m *Manager) ConnectToSource(ctx context.Context, dataHandler func([]byte) error) error {
	server := m.GetCurrentServer()
	if server == nil {
		return fmt.Errorf("no available source server")
	}

	log.Printf("Connecting to source server: %s (%s)", server.Name, server.Address)

	// 尝试连接当前服务器
	conn, err := m.connectWithRetry(ctx, server)
	if err != nil {
		log.Printf("Failed to connect to current server %s: %v", server.Name, err)

		// 记录失败并尝试切换
		m.recordFailure(server.ID)
		if m.shouldFailover(server.ID) {
			m.performFailover()
			// 使用新服务器重试
			return m.ConnectToSource(ctx, dataHandler)
		}
		return err
	}
	defer conn.Close()

	// 重置失败计数
	m.resetFailureCount(server.ID)

	log.Printf("Successfully connected to source server: %s", server.Name)

	// 开始读取数据
	return m.readDataFromSource(ctx, conn, dataHandler)
}

// connectWithRetry 带重试的连接
// 参数: ctx - 上下文, server - 服务器配置
// 返回: 连接对象和错误信息
func (m *Manager) connectWithRetry(ctx context.Context, server *config.SourceServer) (net.Conn, error) {
	var lastErr error

	for attempt := 0; attempt <= server.MaxRetries; attempt++ {
		if attempt > 0 {
			// 重试前等待
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}

		// 尝试连接
		conn, err := net.DialTimeout("tcp", server.Address, server.Timeout)
		if err == nil {
			return conn, nil
		}

		lastErr = err
		log.Printf("Connection attempt %d to %s failed: %v", attempt+1, server.Address, err)
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %v", server.MaxRetries+1, lastErr)
}

// readDataFromSource 从源服务器读取数据
// 参数: ctx - 上下文, conn - 连接, dataHandler - 数据处理函数
// 返回: 错误信息
func (m *Manager) readDataFromSource(ctx context.Context, conn net.Conn, dataHandler func([]byte) error) error {
	buffer := make([]byte, 4096)
	heartbeatTicker := time.NewTicker(3 * time.Second) // 3秒检查一次心跳
	defer heartbeatTicker.Stop()

	// 连接建立后立即发送身份认证包
	if err := m.sendAuthPacket(conn); err != nil {
		log.Printf("Failed to send auth packet: %v", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.shutdownChan:
			return fmt.Errorf("manager shutdown")
		case <-heartbeatTicker.C:
			// 检查是否需要发送心跳
			if m.heartbeatManager.ShouldSendHeartbeat() {
				if err := m.sendHeartbeat(conn); err != nil {
					log.Printf("Failed to send heartbeat: %v", err)
					return err
				}
			}

			// 检查读空闲超时
			if m.heartbeatManager.IsReadIdle() {
				log.Printf("Read idle timeout, closing connection")
				return fmt.Errorf("read idle timeout")
			}
		default:
			// 设置读取超时
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))

			n, err := conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// 超时，继续循环
					continue
				}
				return fmt.Errorf("failed to read from source: %v", err)
			}

			if n > 0 {
				// 处理接收到的数据
				data := make([]byte, n)
				copy(data, buffer[:n])

				// 更新心跳接收时间
				m.heartbeatManager.UpdateHeartbeatReceived()

				// 检查是否是心跳包
				if m.heartbeatManager.IsHeartbeatPacket(data) {
					log.Printf("Received heartbeat packet")
					continue
				}

				// 使用分隔符处理器处理数据，防粘包
				packets, err := m.delimiterHandler.ProcessData(data)
				if err != nil {
					log.Printf("Error processing data with delimiter: %v", err)
					continue
				}

				// 处理每个完整的数据包
				for _, packet := range packets {
					// 验证数据包
					if !m.delimiterHandler.ValidatePacket(packet) {
						log.Printf("Invalid packet received: %d bytes", len(packet))
						continue
					}

					// 使用协议处理器进一步解析
					basePackages, err := m.protocolHandler.ProcessData(packet)
					if err != nil {
						log.Printf("Error parsing base package: %v", err)
						// 如果协议解析失败，直接处理原始数据
						if err := dataHandler(packet); err != nil {
							log.Printf("Error handling raw data: %v", err)
						}
						continue
					}

					// 处理每个解析后的数据包
					for _, pkg := range basePackages {
						if err := dataHandler(pkg.Data); err != nil {
							log.Printf("Error handling data from source: %v", err)
							// 继续处理，不中断连接
						}
					}
				}
			}
		}
	}
}

// healthCheckLoop 健康检查循环
// 参数: ctx - 上下文
func (m *Manager) healthCheckLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // 默认30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.shutdownChan:
			return
		case <-ticker.C:
			m.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck 执行健康检查
// 参数: ctx - 上下文
func (m *Manager) performHealthCheck(ctx context.Context) {
	// 检查主服务器
	if m.config.Primary.Enabled {
		if !m.healthChecker.IsHealthy(ctx, &m.config.Primary) {
			log.Printf("Primary server %s is unhealthy", m.config.Primary.Name)
			m.recordFailure(m.config.Primary.ID)
		} else {
			m.resetFailureCount(m.config.Primary.ID)
		}
	}

	// 检查备用服务器
	if m.config.Backup.Enabled {
		if !m.healthChecker.IsHealthy(ctx, &m.config.Backup) {
			log.Printf("Backup server %s is unhealthy", m.config.Backup.Name)
			m.recordFailure(m.config.Backup.ID)
		} else {
			m.resetFailureCount(m.config.Backup.ID)
		}
	}

	// 检查是否需要故障切换
	m.checkAndPerformFailover()
}

// recordFailure 记录服务器失败
// 参数: serverID - 服务器ID
func (m *Manager) recordFailure(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.failureCounts[serverID]++
	m.lastFailTime[serverID] = time.Now()

	log.Printf("Recorded failure for server %s, total failures: %d",
		serverID, m.failureCounts[serverID])
}

// resetFailureCount 重置服务器失败计数
// 参数: serverID - 服务器ID
func (m *Manager) resetFailureCount(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failureCounts[serverID] > 0 {
		log.Printf("Resetting failure count for server %s", serverID)
		m.failureCounts[serverID] = 0
	}
}

// shouldFailover 判断是否应该进行故障切换
// 参数: serverID - 服务器ID
// 返回: 是否应该切换
func (m *Manager) shouldFailover(serverID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 获取服务器配置
	var server *config.SourceServer
	if serverID == m.config.Primary.ID {
		server = &m.config.Primary
	} else if serverID == m.config.Backup.ID {
		server = &m.config.Backup
	} else {
		return false
	}

	// 检查失败次数是否达到阈值
	return m.failureCounts[serverID] >= server.FailoverThreshold
}

// checkAndPerformFailover 检查并执行故障切换
func (m *Manager) checkAndPerformFailover() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果当前是主服务器且主服务器失败，切换到备用服务器
	if m.currentServer.ID == m.config.Primary.ID &&
		m.failureCounts[m.config.Primary.ID] >= m.config.Primary.FailoverThreshold &&
		m.config.Backup.Enabled {

		log.Printf("Failing over from primary to backup server")
		m.currentServer = &m.config.Backup
		return
	}

	// 如果当前是备用服务器且主服务器恢复，切换回主服务器
	if m.currentServer.ID == m.config.Backup.ID &&
		m.failureCounts[m.config.Primary.ID] == 0 &&
		m.config.Primary.Enabled {

		log.Printf("Failing back from backup to primary server")
		m.currentServer = &m.config.Primary
		return
	}
}

// performFailover 执行故障切换
func (m *Manager) performFailover() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentServer.ID == m.config.Primary.ID && m.config.Backup.Enabled {
		log.Printf("Performing failover: %s -> %s",
			m.config.Primary.Name, m.config.Backup.Name)
		m.currentServer = &m.config.Backup
	} else if m.currentServer.ID == m.config.Backup.ID && m.config.Primary.Enabled {
		log.Printf("Performing failover: %s -> %s",
			m.config.Backup.Name, m.config.Primary.Name)
		m.currentServer = &m.config.Primary
	}
}

// sendAuthPacket 发送身份认证包
// 参数: conn - 连接
// 返回: 错误信息
func (m *Manager) sendAuthPacket(conn net.Conn) error {
	authPacket, err := m.authManager.GenerateAuthPacket()
	if err != nil {
		return fmt.Errorf("failed to generate auth packet: %v", err)
	}

	// 添加分隔符
	packetWithDelimiter := m.delimiterHandler.AddDelimiter(authPacket)

	// 设置写超时
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write(packetWithDelimiter)
	if err != nil {
		return fmt.Errorf("failed to send auth packet: %v", err)
	}

	log.Printf("Sent auth packet to %s: %d bytes", m.GetCurrentServer().Name, len(packetWithDelimiter))
	return nil
}

// sendHeartbeat 发送心跳包
// 参数: conn - 连接
// 返回: 错误信息
func (m *Manager) sendHeartbeat(conn net.Conn) error {
	heartbeatPacket := m.heartbeatManager.GenerateHeartbeatPacket()

	// 设置写超时
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	_, err := conn.Write(heartbeatPacket)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %v", err)
	}

	// 更新心跳发送时间
	m.heartbeatManager.UpdateHeartbeatSent()

	log.Printf("Sent heartbeat packet to %s", m.GetCurrentServer().Name)
	return nil
}

// GetStatus 获取管理器状态
// 返回: 状态信息
func (m *Manager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"is_running":     m.isRunning,
		"current_server": m.currentServer.Name,
		"failure_counts": m.failureCounts,
		"last_fail_time": m.lastFailTime,
	}
}
