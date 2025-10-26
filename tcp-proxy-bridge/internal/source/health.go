// internal/source/health.go
package source

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"tcp-proxy-bridge/internal/config"
)

// HealthChecker 健康检查器
// 负责检查源服务器的健康状态
type HealthChecker struct {
	// 可以添加更多配置，如检查间隔、超时时间等
}

// NewHealthChecker 创建健康检查器
// 返回: 健康检查器实例
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}

// IsHealthy 检查服务器是否健康
// 参数: ctx - 上下文, server - 服务器配置
// 返回: 是否健康
func (hc *HealthChecker) IsHealthy(ctx context.Context, server *config.SourceServer) bool {
	if !server.Enabled {
		return false
	}

	// 尝试建立TCP连接
	conn, err := net.DialTimeout("tcp", server.Address, server.HealthCheckTimeout)
	if err != nil {
		log.Printf("Health check failed for %s (%s): %v", server.Name, server.Address, err)
		return false
	}
	defer conn.Close()

	// 可以在这里添加更复杂的健康检查逻辑
	// 例如发送ping消息、检查响应等

	// 简单的连接测试通过
	log.Printf("Health check passed for %s (%s)", server.Name, server.Address)
	return true
}

// CheckAllServers 检查所有服务器的健康状态
// 参数: ctx - 上下文, servers - 服务器配置列表
// 返回: 健康状态映射
func (hc *HealthChecker) CheckAllServers(ctx context.Context, servers []*config.SourceServer) map[string]bool {
	results := make(map[string]bool)

	for _, server := range servers {
		results[server.ID] = hc.IsHealthy(ctx, server)
	}

	return results
}

// StartPeriodicCheck 启动定期健康检查
// 参数: ctx - 上下文, servers - 服务器配置列表, interval - 检查间隔, callback - 结果回调
func (hc *HealthChecker) StartPeriodicCheck(ctx context.Context, servers []*config.SourceServer,
	interval time.Duration, callback func(map[string]bool)) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			results := hc.CheckAllServers(ctx, servers)
			if callback != nil {
				callback(results)
			}
		}
	}
}

// PingServer 向服务器发送ping消息
// 参数: ctx - 上下文, server - 服务器配置, message - ping消息
// 返回: 响应和错误信息
func (hc *HealthChecker) PingServer(ctx context.Context, server *config.SourceServer, message []byte) ([]byte, error) {
	// 创建带超时的上下文
	pingCtx, cancel := context.WithTimeout(ctx, server.HealthCheckTimeout)
	defer cancel()

	// 建立连接
	conn, err := net.DialTimeout("tcp", server.Address, server.HealthCheckTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", server.Address, err)
	}
	defer conn.Close()

	// 发送ping消息
	if _, err := conn.Write(message); err != nil {
		return nil, fmt.Errorf("failed to send ping to %s: %v", server.Address, err)
	}

	// 读取响应
	response := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(server.HealthCheckTimeout))

	n, err := conn.Read(response)
	if err != nil {
		return nil, fmt.Errorf("failed to read ping response from %s: %v", server.Address, err)
	}

	// 检查上下文是否被取消
	select {
	case <-pingCtx.Done():
		return nil, pingCtx.Err()
	default:
	}

	return response[:n], nil
}

// GetServerLatency 获取服务器延迟
// 参数: ctx - 上下文, server - 服务器配置
// 返回: 延迟时间和错误信息
func (hc *HealthChecker) GetServerLatency(ctx context.Context, server *config.SourceServer) (time.Duration, error) {
	start := time.Now()

	// 简单的连接测试
	conn, err := net.DialTimeout("tcp", server.Address, server.HealthCheckTimeout)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	latency := time.Since(start)
	return latency, nil
}
