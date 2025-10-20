// internal/metrics/essential.go
package metrics

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

// EssentialMetrics 核心指标收集器
// 只收集最关键的4个指标，用于系统监控和故障排查
type EssentialMetrics struct {
	// 使用原子操作保证并发安全

	// MessagesReceived 接收到的消息总数
	// 用途：监控系统输入流量，了解负载情况
	MessagesReceived atomic.Int64

	// MessagesForwarded 成功转发的消息总数
	// 用途：监控系统输出流量，了解处理能力
	MessagesForwarded atomic.Int64

	// MessageErrors 转发失败的消息总数
	// 用途：监控系统健康状态，错误率异常时告警
	MessageErrors atomic.Int64

	// ActiveConnections 当前活跃的TCP连接数
	// 用途：监控系统并发负载，防止连接泄漏
	ActiveConnections atomic.Int32
}

// 全局指标实例
var globalMetrics = &EssentialMetrics{}

// IncMessagesReceived 增加接收消息计数
// 在TCP服务器成功接收并保存消息后调用
func IncMessagesReceived() {
	globalMetrics.MessagesReceived.Add(1)
}

// IncMessagesForwarded 增加成功转发计数
// 在消息成功发送到目标服务器后调用
func IncMessagesForwarded() {
	globalMetrics.MessagesForwarded.Add(1)
}

// IncMessageErrors 增加错误计数
// 在消息转发失败时调用
func IncMessageErrors() {
	globalMetrics.MessageErrors.Add(1)
}

// IncActiveConnections 增加活跃连接数
// 在新TCP连接建立时调用
func IncActiveConnections() {
	globalMetrics.ActiveConnections.Add(1)
}

// DecActiveConnections 减少活跃连接数
// 在TCP连接关闭时调用
func DecActiveConnections() {
	globalMetrics.ActiveConnections.Add(-1)
}

// GetMetricsSnapshot 获取指标快照
// 返回: 包含所有当前指标值的map
// 用途：定期日志记录、健康检查、调试信息
func GetMetricsSnapshot() map[string]interface{} {
	return map[string]interface{}{
		"messages_received":  globalMetrics.MessagesReceived.Load(),
		"messages_forwarded": globalMetrics.MessagesForwarded.Load(),
		"message_errors":     globalMetrics.MessageErrors.Load(),
		"active_connections": globalMetrics.ActiveConnections.Load(),
		"timestamp":          time.Now().Format(time.RFC3339),
	}
}

// GetMetricsSummary 获取指标摘要字符串
// 返回: 格式化的指标摘要
// 用途：日志输出、状态显示
func GetMetricsSummary() string {
	snapshot := GetMetricsSnapshot()
	return fmt.Sprintf("Received: %d, Forwarded: %d, Errors: %d, Connections: %d",
		snapshot["messages_received"],
		snapshot["messages_forwarded"],
		snapshot["message_errors"],
		snapshot["active_connections"])
}

// LogMetrics 记录指标到日志
// 用途：定期输出系统状态，便于监控
func LogMetrics() {
	summary := GetMetricsSummary()
	log.Printf("SYSTEM METRICS - %s", summary)
}

// CalculateErrorRate 计算错误率
// 返回: 错误率百分比 (0.0 - 1.0)
// 用途：健康检查、自动告警
func CalculateErrorRate() float64 {
	received := globalMetrics.MessagesReceived.Load()
	errors := globalMetrics.MessageErrors.Load()

	if received == 0 {
		return 0.0
	}

	return float64(errors) / float64(received)
}

// Reset 重置所有指标（主要用于测试）
func Reset() {
	globalMetrics.MessagesReceived.Store(0)
	globalMetrics.MessagesForwarded.Store(0)
	globalMetrics.MessageErrors.Store(0)
	globalMetrics.ActiveConnections.Store(0)
}

// GetConnectionCount 获取当前连接数
// 返回: 活跃连接数量
// 用途：负载监控、资源管理
func GetConnectionCount() int32 {
	return globalMetrics.ActiveConnections.Load()
}

// GetTotalProcessed 获取总处理消息数
// 返回: 接收到的消息总数
// 用途：性能统计、容量规划
func GetTotalProcessed() int64 {
	return globalMetrics.MessagesReceived.Load()
}
