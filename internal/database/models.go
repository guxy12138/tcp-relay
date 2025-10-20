// internal/database/models.go
package database

import "time"

// Message 消息数据模型
// 对应message_queue表，存储接收到的TCP消息
type Message struct {
	ID           int64      `db:"id"`            // 消息ID，主键
	SourceIP     string     `db:"source_ip"`     // 来源IP地址
	OriginalData []byte     `db:"original_data"` // 原始消息数据
	DataLength   int        `db:"data_length"`   // 数据长度
	CreatedAt    time.Time  `db:"created_at"`    // 创建时间
	ProcessedAt  *time.Time `db:"processed_at"`  // 处理完成时间
	Status       string     `db:"status"`        // 消息状态
}

// TargetDeliveryStatus 目标投递状态模型
// 对应target_delivery_status表，记录消息到每个目标的投递状态
type TargetDeliveryStatus struct {
	ID               int64      `db:"id"`                 // 投递状态ID
	MessageID        int64      `db:"message_id"`         // 关联的消息ID
	TargetServerID   string     `db:"target_server_id"`   // 目标服务器ID
	TargetServerName string     `db:"target_server_name"` // 目标服务器名称
	TargetAddress    string     `db:"target_address"`     // 目标服务器地址
	Status           string     `db:"status"`             // 投递状态
	SendAttempts     int        `db:"send_attempts"`      // 发送尝试次数
	MaxAttempts      int        `db:"max_attempts"`       // 最大尝试次数
	LastAttemptAt    *time.Time `db:"last_attempt_at"`    // 最后尝试时间
	NextRetryAt      *time.Time `db:"next_retry_at"`      // 下次重试时间
	SentAt           *time.Time `db:"sent_at"`            // 成功发送时间
	LastError        *string    `db:"last_error"`         // 最后错误信息
	DataSize         int        `db:"data_size"`          // 数据大小
	CreatedAt        time.Time  `db:"created_at"`         // 创建时间
	UpdatedAt        time.Time  `db:"updated_at"`         // 更新时间
}

// TargetServer 目标服务器配置模型
// 对应target_servers表，存储目标服务器信息
type TargetServer struct {
	ID                string        `db:"id"`                // 服务器ID
	Name              string        `db:"name"`              // 服务器名称
	Address           string        `db:"address"`           // 服务器地址
	Enabled           bool          `db:"enabled"`           // 是否启用
	IsOnline          bool          `db:"is_online"`         // 是否在线
	LastHealthCheck   *time.Time    `db:"last_health_check"` // 最后健康检查时间
	Timeout           time.Duration // 连接超时时间
	MaxRetries        int           `db:"max_retries"`         // 最大重试次数
	BatchSize         int           `db:"batch_size"`          // 批量大小
	Priority          int           `db:"priority"`            // 新增优先级字段
	TotalMessagesSent int64         `db:"total_messages_sent"` // 总发送消息数
	TotalErrors       int64         `db:"total_errors"`        // 总错误数
	LastSuccessAt     *time.Time    `db:"last_success_at"`     // 最后成功时间
}

// 消息状态常量定义
const (
	StatusReceived = "received" // 已接收 - 消息已保存到数据库
	StatusPending  = "pending"  // 等待发送 - 消息等待发送到目标服务器
	StatusSending  = "sending"  // 发送中 - 消息正在发送过程中
	StatusSent     = "sent"     // 已发送 - 消息成功发送到目标服务器
	StatusFailed   = "failed"   // 发送失败 - 消息发送失败
)
