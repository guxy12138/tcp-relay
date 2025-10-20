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
	Server        ServerConfig    `yaml:"server"`
	Database      DatabaseConfig  `yaml:"database"`
	Forwarder     ForwarderConfig `yaml:"forwarder"`
	TargetServers []TargetServer  `yaml:"target_servers"` // 目标服务器配置
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
