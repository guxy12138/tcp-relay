// internal/database/synchronizer.go
package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"tcp-proxy-bridge/internal/config"
)

// TargetSynchronizer 目标服务器同步器
// 负责将配置文件中的目标服务器配置同步到数据库
type TargetSynchronizer struct {
	db *Postgres
}

// NewTargetSynchronizer 创建目标服务器同步器
// 参数: db - 数据库实例
// 返回: 同步器实例
func NewTargetSynchronizer(db *Postgres) *TargetSynchronizer {
	return &TargetSynchronizer{
		db: db,
	}
}

// SyncTargetServers 同步目标服务器配置到数据库
// 参数: servers - 配置文件中定义的目标服务器列表
// 返回: 错误信息
func (s *TargetSynchronizer) SyncTargetServers(servers []config.TargetServer) error {
	log.Printf("Starting target servers synchronization, %d servers configured", len(servers))

	// 开始数据库事务
	tx, err := s.db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 获取数据库中现有的目标服务器
	existingServers, err := s.getExistingServers(tx)
	if err != nil {
		return fmt.Errorf("failed to get existing servers: %v", err)
	}

	// 同步每个服务器配置
	for _, server := range servers {
		if err := s.syncSingleServer(tx, server, existingServers); err != nil {
			return fmt.Errorf("failed to sync server %s: %v", server.ID, err)
		}
	}

	// 禁用配置文件中不存在的服务器（软删除）
	if err := s.disableMissingServers(tx, servers, existingServers); err != nil {
		return fmt.Errorf("failed to disable missing servers: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Target servers synchronization completed successfully")
	return nil
}

// getExistingServers 获取数据库中现有的目标服务器
// 参数: tx - 数据库事务
// 返回: 服务器ID到服务器信息的映射
func (s *TargetSynchronizer) getExistingServers(tx *sql.Tx) (map[string]*TargetServer, error) {
	query := `SELECT id, name, address, enabled, connection_timeout_sec, max_retries, batch_size, priority 
              FROM target_servers`

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := make(map[string]*TargetServer)
	for rows.Next() {
		var server TargetServer
		var timeoutSec int

		err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.Address,
			&server.Enabled,
			&timeoutSec,
			&server.MaxRetries,
			&server.BatchSize,
			&server.Priority,
		)
		if err != nil {
			return nil, err
		}

		server.Timeout = time.Duration(timeoutSec) * time.Second
		servers[server.ID] = &server
	}

	return servers, nil
}

// syncSingleServer 同步单个服务器配置
// 参数: tx - 数据库事务, server - 服务器配置, existingServers - 现有服务器映射
// 返回: 错误信息
func (s *TargetSynchronizer) syncSingleServer(tx *sql.Tx, server config.TargetServer, existingServers map[string]*TargetServer) error {
	if existing, exists := existingServers[server.ID]; exists {
		// 服务器已存在，检查是否需要更新
		if s.needUpdate(existing, server) {
			return s.updateServer(tx, server)
		}
		log.Printf("Target server %s already up to date", server.ID)
	} else {
		// 服务器不存在，创建新记录
		return s.createServer(tx, server)
	}
	return nil
}

// needUpdate 检查服务器配置是否需要更新
// 参数: existing - 现有服务器配置, new - 新服务器配置
// 返回: 是否需要更新
func (s *TargetSynchronizer) needUpdate(existing *TargetServer, new config.TargetServer) bool {
	return existing.Name != new.Name ||
		existing.Address != new.Address ||
		existing.Enabled != new.Enabled ||
		existing.Timeout != new.Timeout ||
		existing.MaxRetries != new.MaxRetries ||
		existing.BatchSize != new.BatchSize ||
		existing.Priority != new.Priority
}

// updateServer 更新现有服务器配置
// 参数: tx - 数据库事务, server - 服务器配置
// 返回: 错误信息
func (s *TargetSynchronizer) updateServer(tx *sql.Tx, server config.TargetServer) error {
	query := `UPDATE target_servers 
              SET name = $1, address = $2, enabled = $3, 
                  connection_timeout_sec = $4, max_retries = $5, 
                  batch_size = $6, priority = $7
              WHERE id = $8`

	_, err := tx.Exec(
		query,
		server.Name,
		server.Address,
		server.Enabled,
		int(server.Timeout.Seconds()),
		server.MaxRetries,
		server.BatchSize,
		server.Priority,
		server.ID,
	)

	if err != nil {
		return err
	}

	log.Printf("Updated target server: %s (%s)", server.Name, server.Address)
	return nil
}

// createServer 创建新的服务器配置
// 参数: tx - 数据库事务, server - 服务器配置
// 返回: 错误信息
func (s *TargetSynchronizer) createServer(tx *sql.Tx, server config.TargetServer) error {
	query := `INSERT INTO target_servers 
              (id, name, address, enabled, connection_timeout_sec, max_retries, batch_size, priority)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := tx.Exec(
		query,
		server.ID,
		server.Name,
		server.Address,
		server.Enabled,
		int(server.Timeout.Seconds()),
		server.MaxRetries,
		server.BatchSize,
		server.Priority,
	)

	if err != nil {
		return err
	}

	log.Printf("Created target server: %s (%s)", server.Name, server.Address)
	return nil
}

// disableMissingServers 禁用配置文件中不存在的服务器
// 参数: tx - 数据库事务, configuredServers - 配置的服务器列表, existingServers - 现有服务器映射
// 返回: 错误信息
func (s *TargetSynchronizer) disableMissingServers(tx *sql.Tx, configuredServers []config.TargetServer, existingServers map[string]*TargetServer) error {
	// 创建配置文件中服务器ID的集合
	configuredIDs := make(map[string]bool)
	for _, server := range configuredServers {
		configuredIDs[server.ID] = true
	}

	// 禁用不在配置文件中的服务器
	for id := range existingServers {
		if !configuredIDs[id] {
			query := `UPDATE target_servers SET enabled = false WHERE id = $1`
			if _, err := tx.Exec(query, id); err != nil {
				return err
			}
			log.Printf("Disabled target server (not in configuration): %s", id)
		}
	}

	return nil
}

// GetEnabledTargetServers 获取所有启用的目标服务器（从数据库读取）
// 返回: 启用的服务器列表和错误信息
func (s *TargetSynchronizer) GetEnabledTargetServers() ([]*TargetServer, error) {
	query := `SELECT id, name, address, enabled, connection_timeout_sec, max_retries, batch_size, priority
              FROM target_servers 
              WHERE enabled = true 
              ORDER BY priority ASC, created_at ASC`

	rows, err := s.db.db.Query(query)
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
			&server.Enabled,
			&timeoutSec,
			&server.MaxRetries,
			&server.BatchSize,
			&server.Priority,
		)
		if err != nil {
			return nil, err
		}

		server.Timeout = time.Duration(timeoutSec) * time.Second
		servers = append(servers, &server)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("Loaded %d enabled target servers from database", len(servers))
	return servers, nil
}
