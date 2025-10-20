// internal/health/minimal.go
package health

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// MinimalServer 最小化健康检查服务器
// 只检查数据库连接和TCP端口监听状态
type MinimalServer struct {
	server  *http.Server // HTTP服务器实例
	db      *sql.DB      // 数据库连接
	tcpPort int          // TCP服务端口
}

// NewMinimalServer 创建最小化健康检查服务器
// 参数: healthPort - 健康检查服务端口, tcpPort - TCP服务端口, db - 数据库连接
// 返回: 健康检查服务器实例
func NewMinimalServer(healthPort, tcpPort int, db *sql.DB) *MinimalServer {
	// 创建HTTP路由
	mux := http.NewServeMux()

	server := &MinimalServer{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", healthPort),
			Handler: mux,
		},
		db:      db,
		tcpPort: tcpPort,
	}

	// 注册健康检查端点
	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("/ready", server.readyHandler)

	return server
}

// Start 启动健康检查服务器
// 返回: 错误信息
func (s *MinimalServer) Start() error {
	log.Printf("Minimal health server starting on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop 停止健康检查服务器
// 参数: ctx - 上下文
// 返回: 错误信息
func (s *MinimalServer) Stop(ctx context.Context) error {
	log.Println("Stopping health server...")
	return s.server.Shutdown(ctx)
}

// healthHandler 健康检查处理器
// 检查数据库连接和TCP端口监听状态
func (s *MinimalServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 检查数据库连接
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		log.Printf("Health check failed: database unavailable - %v", err)
		http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
		return
	}

	// 2. 检查TCP服务端口是否在监听
	if !s.checkTCPPort() {
		log.Printf("Health check failed: TCP port %d not listening", s.tcpPort)
		http.Error(w, "TCP port not listening", http.StatusServiceUnavailable)
		return
	}

	// 所有检查通过，返回健康状态
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Printf("Health check passed at %v", time.Now().Format(time.RFC3339))
}

// readyHandler 就绪检查处理器
// 简单的就绪检查，总是返回就绪状态（因为健康检查已经包含了关键依赖检查）
func (s *MinimalServer) readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

// checkTCPPort 检查TCP端口是否在监听
// 返回: 端口是否可连接
func (s *MinimalServer) checkTCPPort() bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", s.tcpPort), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetHealthStatus 获取健康状态（供程序内部使用）
// 返回: 健康状态和错误信息
func (s *MinimalServer) GetHealthStatus() (bool, error) {
	// 检查数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return false, fmt.Errorf("database unavailable: %v", err)
	}

	// 检查TCP端口
	if !s.checkTCPPort() {
		return false, fmt.Errorf("TCP port %d not listening", s.tcpPort)
	}

	return true, nil
}
