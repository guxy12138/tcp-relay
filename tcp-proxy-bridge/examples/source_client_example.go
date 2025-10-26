// examples/source_client_example.go
package main

import (
	"context"
	"log"
	"time"

	"tcp-proxy-bridge/internal/config"
	"tcp-proxy-bridge/internal/source"
)

func main() {
	// 创建源服务器配置
	sourceConfig := &config.SourceServers{
		Primary: config.SourceServer{
			ID:                  "primary-server",
			Name:                "主源服务器",
			Address:             "127.0.0.1:8888",
			Enabled:             true,
			Timeout:             10 * time.Second,
			MaxRetries:          3,
			BatchSize:           100,
			HealthCheckInterval: 30 * time.Second,
			HealthCheckTimeout:  5 * time.Second,
			FailoverThreshold:   3,
		},
		Backup: config.SourceServer{
			ID:                  "backup-server",
			Name:                "备用源服务器",
			Address:             "127.0.0.1:8889",
			Enabled:             true,
			Timeout:             10 * time.Second,
			MaxRetries:          3,
			BatchSize:           50,
			HealthCheckInterval: 30 * time.Second,
			HealthCheckTimeout:  5 * time.Second,
			FailoverThreshold:   3,
		},
	}

	// 创建源服务器管理器
	manager := source.NewManager(sourceConfig)

	// 设置数据处理回调
	manager.SetDataHandler(func(data []byte) error {
		log.Printf("Received data: %d bytes", len(data))
		// 这里可以添加你的数据处理逻辑
		// 例如：解析数据、存储到数据库、转发到其他服务等
		return nil
	})

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动管理器
	if err := manager.Start(ctx); err != nil {
		log.Fatalf("Failed to start source manager: %v", err)
	}

	log.Println("Source server manager started successfully")

	// 连接到源服务器并开始接收数据
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := manager.ConnectToSource(ctx, func(data []byte) error {
					log.Printf("Processing data from source: %d bytes", len(data))
					return nil
				}); err != nil {
					log.Printf("Connection error: %v", err)
					time.Sleep(5 * time.Second) // 等待5秒后重试
				}
			}
		}
	}()

	// 定期打印状态
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")
			manager.Stop(ctx)
			return
		case <-ticker.C:
			status := manager.GetStatus()
			log.Printf("Manager status: %+v", status)
		}
	}
}
