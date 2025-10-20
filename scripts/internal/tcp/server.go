// internal/tcp/server.go
package tcp

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

// Server TCP服务器实现
// 负责监听TCP端口，接收客户端连接并处理数据
type Server struct {
	config      *config.ServerConfig // 服务器配置
	db          *database.Postgres   // 数据库实例
	listener    net.Listener         // TCP监听器
	wg          sync.WaitGroup       // 等待组，用于优雅关闭
	mu          sync.RWMutex         // 读写锁，保护共享状态
	isRunning   bool                 // 服务器运行状态
	connections map[net.Conn]bool    // 活跃连接集合
}

// NewServer 创建新的TCP服务器实例
// 参数: cfg - 服务器配置, db - 数据库实例
// 返回: TCP服务器实例
func NewServer(cfg *config.ServerConfig, db *database.Postgres) *Server {
	return &Server{
		config:      cfg,
		db:          db,
		isRunning:   false,
		connections: make(map[net.Conn]bool),
	}
}

// Start 启动TCP服务器
// 参数: ctx - 上下文，用于控制服务器生命周期
// 返回: 错误信息
func (s *Server) Start(ctx context.Context) error {
	var err error
	// 构建监听地址
	addr := fmt.Sprintf(":%d", s.config.TCPListenPort)

	// 创建TCP监听器
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start TCP server on port %d: %v", s.config.TCPListenPort, err)
	}

	// 设置服务器运行状态
	s.mu.Lock()
	s.isRunning = true
	s.mu.Unlock()

	log.Printf("TCP server started successfully, listening on port %d", s.config.TCPListenPort)

	// 启动连接清理器
	go s.connectionCleaner(ctx)

	// 主接受循环
	for {
		select {
		case <-ctx.Done():
			// 上下文被取消，优雅关闭
			log.Println("TCP server context cancelled, shutting down")
			return nil
		default:
			// 接受新连接
			conn, err := s.listener.Accept()
			if err != nil {
				s.mu.RLock()
				running := s.isRunning
				s.mu.RUnlock()

				if !running {
					// 服务器已停止，正常退出
					return nil
				}
				// 接受连接时出错，记录日志但继续运行
				log.Printf("Error accepting connection: %v", err)
				continue
			}

			// 记录新连接并启动处理协程
			s.mu.Lock()
			s.connections[conn] = true
			s.mu.Unlock()

			// 更新活跃连接指标
			metrics.IncActiveConnections()

			s.wg.Add(1)
			go s.handleConnection(ctx, conn)
		}
	}
}

// handleConnection 处理单个TCP连接
// 参数: ctx - 上下文, conn - TCP连接
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	// 确保连接最终被关闭并从连接集合中移除
	defer func() {
		s.mu.Lock()
		delete(s.connections, conn)
		s.mu.Unlock()
		conn.Close()
		s.wg.Done()
		metrics.DecActiveConnections()
	}()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("New TCP connection established from: %s", remoteAddr)

	// 创建读取缓冲区
	buffer := make([]byte, s.config.MaxMessageSize)

	for {
		select {
		case <-ctx.Done():
			// 上下文被取消，退出处理循环
			log.Printf("Connection context cancelled for: %s", remoteAddr)
			return
		default:
			// 设置读取超时，避免永久阻塞
			conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))

			// 读取客户端数据
			n, err := conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// 读取超时，继续等待新数据
					continue
				}
				// 连接错误或关闭，退出处理循环
				log.Printf("Connection read error from %s: %v", remoteAddr, err)
				return
			}

			if n > 0 {
				// 复制接收到的数据
				data := make([]byte, n)
				copy(data, buffer[:n])

				// 处理接收到的数据
				if err := s.processReceivedData(remoteAddr, data); err != nil {
					log.Printf("Failed to process data from %s: %v", remoteAddr, err)
					// 注意：这里不增加错误计数，因为数据库保存失败已经在processReceivedData中记录了
				} else {
					log.Printf("Successfully processed %d bytes from %s", n, remoteAddr)
					// 增加成功接收消息计数
					metrics.IncMessagesReceived()
				}
			}
		}
	}
}

// processReceivedData 处理接收到的TCP数据
// 参数: sourceIP - 数据来源IP, data - 接收到的数据
// 返回: 错误信息
func (s *Server) processReceivedData(sourceIP string, data []byte) error {
	// 创建消息对象
	message := &database.Message{
		SourceIP:     sourceIP,
		OriginalData: data,
		DataLength:   len(data),
		Status:       database.StatusReceived,
	}

	// 保存消息到数据库
	if err := s.db.SaveMessage(message); err != nil {
		// 数据库保存失败，增加错误计数
		metrics.IncMessageErrors()
		return fmt.Errorf("failed to save message from %s to database: %v", sourceIP, err)
	}

	log.Printf("Successfully saved message %d from %s to database", message.ID, sourceIP)
	return nil
}

// connectionCleaner 连接清理器，定期检查并清理失效连接
// 参数: ctx - 上下文
func (s *Server) connectionCleaner(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanStaleConnections()
		}
	}
}

// cleanStaleConnections 清理失效的连接
func (s *Server) cleanStaleConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanedCount := 0
	for conn := range s.connections {
		// 尝试发送空数据来检测连接是否仍然活跃
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if _, err := conn.Write([]byte{}); err != nil {
			// 连接已失效，关闭并移除
			conn.Close()
			delete(s.connections, conn)
			cleanedCount++
			metrics.DecActiveConnections()
		}
	}

	if cleanedCount > 0 {
		log.Printf("Cleaned up %d stale connections", cleanedCount)
	}
}

// Stop 停止TCP服务器
// 参数: ctx - 上下文，用于控制关闭超时
func (s *Server) Stop(ctx context.Context) {
	// 设置服务器状态为停止
	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()

	log.Println("Stopping TCP server...")

	// 关闭监听器，停止接受新连接
	if s.listener != nil {
		s.listener.Close()
	}

	// 关闭所有活跃连接
	s.mu.Lock()
	for conn := range s.connections {
		conn.Close()
		delete(s.connections, conn)
		metrics.DecActiveConnections()
	}
	s.mu.Unlock()

	// 等待所有连接处理协程完成
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// 等待所有处理完成或超时
	select {
	case <-done:
		log.Println("All TCP connections closed gracefully")
	case <-ctx.Done():
		log.Println("Force shutdown TCP server due to timeout")
	}
}

// GetConnectionCount 获取当前活跃连接数
// 返回: 活跃连接数量
func (s *Server) GetConnectionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connections)
}
