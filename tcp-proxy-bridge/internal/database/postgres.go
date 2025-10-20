// internal/database/postgres.go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"tcp-proxy-bridge/internal/config"

	_ "github.com/lib/pq" // PostgreSQL驱动
)

// Postgres 数据库操作封装
// 提供对PostgreSQL数据库的CRUD操作
type Postgres struct {
	db *sql.DB // 数据库连接实例
}

func (p *Postgres) DB() *sql.DB { return p.db }

// NewPostgres 创建新的数据库连接
// 参数: cfg - 数据库配置信息
// 返回: 数据库实例和错误信息
func NewPostgres(cfg config.DatabaseConfig) (*Postgres, error) {
	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)

	// 打开数据库连接
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 配置连接池参数
	db.SetMaxOpenConns(cfg.MaxOpenConns)       // 最大打开连接数
	db.SetMaxIdleConns(cfg.MaxIdleConns)       // 最大空闲连接数
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime) // 连接最大生命周期

	// 测试数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return &Postgres{db: db}, nil
}

// SaveMessage 保存接收到的消息到数据库
// 同时为每个启用的目标服务器创建投递状态记录
// 参数: msg - 要保存的消息对象
// 返回: 错误信息
func (p *Postgres) SaveMessage(msg *Message) error {
	// SQL插入语句，返回生成的ID和创建时间
	query := `INSERT INTO message_queue (source_ip, original_data, data_length, status) 
              VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	// 执行插入操作
	err := p.db.QueryRow(query, msg.SourceIP, msg.OriginalData, msg.DataLength, msg.Status).
		Scan(&msg.ID, &msg.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to save message: %v", err)
	}

	log.Printf("Successfully saved message %d from %s", msg.ID, msg.SourceIP)

	// 获取所有启用的目标服务器
	targets, err := p.GetEnabledTargetServers()
	if err != nil {
		return fmt.Errorf("failed to get target servers: %v", err)
	}

	// 为每个目标服务器创建投递状态记录
	for _, target := range targets {
		delivery := &TargetDeliveryStatus{
			MessageID:        msg.ID,
			TargetServerID:   target.ID,
			TargetServerName: target.Name,
			TargetAddress:    target.Address,
			Status:           StatusPending,
			MaxAttempts:      target.MaxRetries,
			DataSize:         msg.DataLength,
		}

		if err := p.CreateDeliveryStatus(delivery); err != nil {
			log.Printf("Failed to create delivery status for target %s: %v", target.ID, err)
		}
	}

	return nil
}

// GetPendingMessagesForTarget 获取指定目标服务器的待处理消息
// 参数: targetID - 目标服务器ID, limit - 最大返回数量
// 返回: 消息列表和错误信息
func (p *Postgres) GetPendingMessagesForTarget(targetID string, limit int) ([]*Message, error) {
	// SQL查询语句，联合查询消息表和投递状态表
	query := `
        SELECT mq.id, mq.source_ip, mq.original_data, mq.data_length, 
               mq.created_at, mq.status
        FROM message_queue mq
        JOIN target_delivery_status tds ON mq.id = tds.message_id
        WHERE tds.target_server_id = $1 
          AND tds.status IN ('pending', 'failed')  -- 待处理或失败的消息
          AND (tds.next_retry_at IS NULL OR tds.next_retry_at <= NOW())  -- 可重试的消息
          AND tds.send_attempts < tds.max_attempts  -- 未超过最大尝试次数
        ORDER BY mq.created_at ASC  -- 按创建时间排序，先处理旧消息
        LIMIT $2`

	// 执行查询
	rows, err := p.db.Query(query, targetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 解析查询结果
	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SourceIP, &msg.OriginalData, &msg.DataLength, &msg.CreatedAt, &msg.Status)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	// 检查遍历过程中是否发生错误
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// UpdateDeliveryStatus 更新消息投递状态
// 参数: messageID - 消息ID, targetID - 目标服务器ID, status - 新状态,
//
//	attempts - 尝试次数, errorMsg - 错误信息
//
// 返回: 错误信息
func (p *Postgres) UpdateDeliveryStatus(messageID int64, targetID string, status string, attempts int, errorMsg *string) error {
	var query string
	var err error

	now := time.Now()

	// 根据状态类型构建不同的SQL语句
	switch status {
	case StatusSending:
		// 更新为发送中状态
		query = `UPDATE target_delivery_status 
                 SET status = $1, last_attempt_at = $2, send_attempts = $3 
                 WHERE message_id = $4 AND target_server_id = $5`
		_, err = p.db.Exec(query, status, now, attempts, messageID, targetID)

	case StatusSent:
		// 更新为已发送状态
		query = `UPDATE target_delivery_status 
                 SET status = $1, sent_at = $2, send_attempts = $3, last_error = NULL 
                 WHERE message_id = $4 AND target_server_id = $5`
		_, err = p.db.Exec(query, status, now, attempts, messageID, targetID)

	case StatusFailed:
		// 更新为失败状态，计算下次重试时间
		nextRetry := p.calculateNextRetryTime(attempts)
		query = `UPDATE target_delivery_status 
                 SET status = $1, last_attempt_at = $2, send_attempts = $3, 
                     next_retry_at = $4, last_error = $5, error_count = error_count + 1
                 WHERE message_id = $6 AND target_server_id = $7`
		_, err = p.db.Exec(query, status, now, attempts, nextRetry, errorMsg, messageID, targetID)

	default:
		return fmt.Errorf("unknown status: %s", status)
	}

	return err
}

// GetEnabledTargetServers 获取所有启用的目标服务器配置
// 返回: 目标服务器列表和错误信息
func (p *Postgres) GetEnabledTargetServers() ([]*TargetServer, error) {
	query := `SELECT id, name, address, max_retries, connection_timeout_sec, batch_size
              FROM target_servers 
              WHERE enabled = true 
              ORDER BY created_at ASC` // 按创建时间排序

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*TargetServer
	for rows.Next() {
		var server TargetServer
		var timeoutSec int

		err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.Address,
			&server.MaxRetries,
			&timeoutSec,
			&server.BatchSize,
		)
		if err != nil {
			return nil, err
		}

		// 转换超时时间为Duration类型
		server.Timeout = time.Duration(timeoutSec) * time.Second
		servers = append(servers, &server)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

// CreateDeliveryStatus 创建投递状态记录
// 参数: delivery - 投递状态信息
// 返回: 错误信息
func (p *Postgres) CreateDeliveryStatus(delivery *TargetDeliveryStatus) error {
	query := `INSERT INTO target_delivery_status 
              (message_id, target_server_id, target_server_name, target_address, 
               status, max_attempts, data_size) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := p.db.Exec(
		query,
		delivery.MessageID,
		delivery.TargetServerID,
		delivery.TargetServerName,
		delivery.TargetAddress,
		delivery.Status,
		delivery.MaxAttempts,
		delivery.DataSize,
	)

	return err
}

// GetDeliveryAttempts 获取消息投递尝试次数
// 参数: messageID - 消息ID, targetID - 目标服务器ID
// 返回: 尝试次数和错误信息
func (p *Postgres) GetDeliveryAttempts(messageID int64, targetID string) (int, error) {
	query := `SELECT send_attempts 
              FROM target_delivery_status 
              WHERE message_id = $1 AND target_server_id = $2`

	var attempts int
	err := p.db.QueryRow(query, messageID, targetID).Scan(&attempts)
	if err != nil {
		return 0, err
	}

	return attempts, nil
}

// calculateNextRetryTime 计算下次重试时间（指数退避算法）
// 参数: attempts - 已尝试次数
// 返回: 下次重试时间
func (p *Postgres) calculateNextRetryTime(attempts int) time.Time {
	// 指数退避: 1分钟, 2分钟, 4分钟, 8分钟...最大60分钟
	delay := time.Duration(1<<uint(attempts)) * time.Minute
	if delay > 60*time.Minute {
		delay = 60 * time.Minute
	}
	return time.Now().Add(delay)
}

// PingContext 检查数据库连接状态
// 参数: ctx - 上下文
// 返回: 错误信息
func (p *Postgres) PingContext(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Close 关闭数据库连接
// 返回: 错误信息
func (p *Postgres) Close() error {
	return p.db.Close()
}
