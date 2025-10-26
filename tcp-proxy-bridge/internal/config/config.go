// internal/config/config.go
package config

import (
	"fmt"
	"net"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Config 应用主配置结构
type Config struct {
	Server         ServerConfig    `yaml:"server"`
	Database       DatabaseConfig  `yaml:"database"`
	Forwarder      ForwarderConfig `yaml:"forwarder"`
	SourceServers  SourceServers   `yaml:"source_servers"` // 源服务器配置（主备）
	Authentication AuthConfig      `yaml:"authentication"` // 身份认证配置
	Heartbeat      HeartbeatConfig `yaml:"heartbeat"`      // 心跳配置
	Delimiter      DelimiterConfig `yaml:"delimiter"`      // 分隔符配置
	TargetServers  []TargetServer  `yaml:"target_servers"` // 目标服务器配置
}

// ServerConfig 服务器相关配置
type ServerConfig struct {
	TCPListenPort   int           `yaml:"tcp_listen_port"`
	HealthCheckPort int           `yaml:"health_check_port"`
	MaxConnections  int           `yaml:"max_connections"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	MaxMessageSize  int           `yaml:"max_message_size"`
}

// DatabaseConfig 数据库连接配置
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Name            string        `yaml:"name"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// ForwarderConfig 消息转发配置
type ForwarderConfig struct {
	BatchSize            int           `yaml:"batch_size"`
	ProcessInterval      time.Duration `yaml:"process_interval"`
	MaxProcessingWorkers int           `yaml:"max_processing_workers"`
	BaseRetryInterval    time.Duration `yaml:"base_retry_interval"`
	MaxRetryInterval     time.Duration `yaml:"max_retry_interval"`
}

// SourceServers 源服务器配置（主备）
type SourceServers struct {
	Primary SourceServer `yaml:"primary"` // 主服务器
	Backup  SourceServer `yaml:"backup"`  // 备用服务器
}

// SourceServer 源服务器配置
type SourceServer struct {
	ID                  string        `yaml:"id"`                    // 服务器唯一标识
	Name                string        `yaml:"name"`                  // 服务器显示名称
	Address             string        `yaml:"address"`               // 服务器地址 (IP:Port)
	Enabled             bool          `yaml:"enabled"`               // 是否启用该服务器
	Timeout             time.Duration `yaml:"timeout"`               // 连接超时时间
	MaxRetries          int           `yaml:"max_retries"`           // 最大重试次数
	BatchSize           int           `yaml:"batch_size"`            // 批量处理大小
	HealthCheckInterval time.Duration `yaml:"health_check_interval"` // 健康检查间隔
	HealthCheckTimeout  time.Duration `yaml:"health_check_timeout"`  // 健康检查超时
	FailoverThreshold   int           `yaml:"failover_threshold"`    // 故障切换阈值
}

// AuthConfig 身份认证配置
type AuthConfig struct {
	Token          string        `yaml:"token"`           // 认证令牌
	SourceID       uint32        `yaml:"source_id"`       // 信源ID
	HostID         uint32        `yaml:"host_id"`         // 信宿ID
	ReauthInterval time.Duration `yaml:"reauth_interval"` // 重新认证间隔
}

// HeartbeatConfig 心跳配置
type HeartbeatConfig struct {
	Interval         time.Duration `yaml:"interval"`           // 心跳发送间隔
	WriteIdleTimeout time.Duration `yaml:"write_idle_timeout"` // 写空闲超时
	ReadIdleTimeout  time.Duration `yaml:"read_idle_timeout"`  // 读空闲超时
}

// DelimiterConfig 分隔符配置
type DelimiterConfig struct {
	Separator       string `yaml:"separator"`         // 分隔符 (十六进制字符串)
	MaxPacketLength int    `yaml:"max_packet_length"` // 最大数据包长度
}

// TargetServer 目标服务器配置
type TargetServer struct {
	ID         string        `yaml:"id"`          // 服务器唯一标识
	Name       string        `yaml:"name"`        // 服务器显示名称
	Address    string        `yaml:"address"`     // 服务器地址 (IP:Port)
	Enabled    bool          `yaml:"enabled"`     // 是否启用该服务器
	Timeout    time.Duration `yaml:"timeout"`     // 连接超时时间
	MaxRetries int           `yaml:"max_retries"` // 最大重试次数
	BatchSize  int           `yaml:"batch_size"`  // 批量处理大小
	Priority   int           `yaml:"priority"`    // 优先级 (数字越小优先级越高)
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ValidateTargetServers 验证目标服务器配置
// 返回: 验证错误信息
func (c *Config) ValidateTargetServers() error {
	if len(c.TargetServers) == 0 {
		return fmt.Errorf("no target servers configured")
	}

	seenIDs := make(map[string]bool)
	for i, server := range c.TargetServers {
		// 检查ID是否重复
		if seenIDs[server.ID] {
			return fmt.Errorf("duplicate target server ID: %s", server.ID)
		}
		seenIDs[server.ID] = true

		// 验证必填字段
		if server.ID == "" {
			return fmt.Errorf("target server %d: ID is required", i)
		}
		if server.Name == "" {
			return fmt.Errorf("target server %s: name is required", server.ID)
		}
		if server.Address == "" {
			return fmt.Errorf("target server %s: address is required", server.ID)
		}

		// 验证地址格式
		if _, _, err := net.SplitHostPort(server.Address); err != nil {
			return fmt.Errorf("target server %s: invalid address format '%s', expected 'host:port'",
				server.ID, server.Address)
		}

		// 验证超时时间
		if server.Timeout <= 0 {
			return fmt.Errorf("target server %s: timeout must be positive", server.ID)
		}

		// 验证重试次数
		if server.MaxRetries < 0 {
			return fmt.Errorf("target server %s: max_retries cannot be negative", server.ID)
		}

		// 验证批量大小
		if server.BatchSize <= 0 {
			return fmt.Errorf("target server %s: batch_size must be positive", server.ID)
		}
	}

	return nil
}

// ValidateSourceServers 验证源服务器配置
// 返回: 验证错误信息
func (c *Config) ValidateSourceServers() error {
	// 验证主服务器
	if err := c.validateSingleSourceServer(&c.SourceServers.Primary, "primary"); err != nil {
		return err
	}

	// 验证备用服务器
	if err := c.validateSingleSourceServer(&c.SourceServers.Backup, "backup"); err != nil {
		return err
	}

	// 检查主备服务器地址不能相同
	if c.SourceServers.Primary.Address == c.SourceServers.Backup.Address {
		return fmt.Errorf("primary and backup servers cannot have the same address")
	}

	return nil
}

// validateSingleSourceServer 验证单个源服务器配置
// 参数: server - 源服务器配置, serverType - 服务器类型（primary/backup）
// 返回: 验证错误信息
func (c *Config) validateSingleSourceServer(server *SourceServer, serverType string) error {
	// 验证必填字段
	if server.ID == "" {
		return fmt.Errorf("%s server: ID is required", serverType)
	}
	if server.Name == "" {
		return fmt.Errorf("%s server: name is required", serverType)
	}
	if server.Address == "" {
		return fmt.Errorf("%s server: address is required", serverType)
	}

	// 验证地址格式
	if _, _, err := net.SplitHostPort(server.Address); err != nil {
		return fmt.Errorf("%s server: invalid address format '%s', expected 'host:port'",
			serverType, server.Address)
	}

	// 验证超时时间
	if server.Timeout <= 0 {
		return fmt.Errorf("%s server: timeout must be positive", serverType)
	}

	// 验证重试次数
	if server.MaxRetries < 0 {
		return fmt.Errorf("%s server: max_retries cannot be negative", serverType)
	}

	// 验证批量大小
	if server.BatchSize <= 0 {
		return fmt.Errorf("%s server: batch_size must be positive", serverType)
	}

	// 验证健康检查间隔
	if server.HealthCheckInterval <= 0 {
		return fmt.Errorf("%s server: health_check_interval must be positive", serverType)
	}

	// 验证健康检查超时
	if server.HealthCheckTimeout <= 0 {
		return fmt.Errorf("%s server: health_check_timeout must be positive", serverType)
	}

	// 验证故障切换阈值
	if server.FailoverThreshold <= 0 {
		return fmt.Errorf("%s server: failover_threshold must be positive", serverType)
	}

	return nil
}

// ValidateAuthentication 验证身份认证配置
// 返回: 验证错误信息
func (c *Config) ValidateAuthentication() error {
	// 验证token
	if c.Authentication.Token == "" {
		return fmt.Errorf("authentication token is required")
	}

	// 验证token长度（应该是32字节的十六进制字符串或32字节的ASCII字符串）
	if len(c.Authentication.Token) != 32 && len(c.Authentication.Token) != 64 {
		return fmt.Errorf("authentication token must be 32 bytes (ASCII) or 64 characters (hex)")
	}

	// 验证信源ID
	if c.Authentication.SourceID == 0 {
		return fmt.Errorf("authentication source_id cannot be zero")
	}

	// 验证信宿ID
	if c.Authentication.HostID == 0 {
		return fmt.Errorf("authentication host_id cannot be zero")
	}

	// 验证重新认证间隔
	if c.Authentication.ReauthInterval <= 0 {
		return fmt.Errorf("authentication reauth_interval must be positive")
	}

	return nil
}

// ValidateHeartbeat 验证心跳配置
// 返回: 验证错误信息
func (c *Config) ValidateHeartbeat() error {
	// 验证心跳间隔
	if c.Heartbeat.Interval <= 0 {
		return fmt.Errorf("heartbeat interval must be positive")
	}

	// 验证写空闲超时
	if c.Heartbeat.WriteIdleTimeout <= 0 {
		return fmt.Errorf("heartbeat write_idle_timeout must be positive")
	}

	// 验证读空闲超时
	if c.Heartbeat.ReadIdleTimeout <= 0 {
		return fmt.Errorf("heartbeat read_idle_timeout must be positive")
	}

	return nil
}

// ValidateDelimiter 验证分隔符配置
// 返回: 验证错误信息
func (c *Config) ValidateDelimiter() error {
	// 验证分隔符
	if c.Delimiter.Separator == "" {
		return fmt.Errorf("delimiter separator is required")
	}

	// 验证分隔符长度（应该是偶数，因为是十六进制字符串）
	if len(c.Delimiter.Separator)%2 != 0 {
		return fmt.Errorf("delimiter separator must be even length (hex string)")
	}

	// 验证最大包长度
	if c.Delimiter.MaxPacketLength <= 0 {
		return fmt.Errorf("delimiter max_packet_length must be positive")
	}

	return nil
}
