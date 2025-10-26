// cmd/server/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tcp-proxy-bridge/internal/config"
	"tcp-proxy-bridge/internal/database"
	"tcp-proxy-bridge/internal/forwarder"
	"tcp-proxy-bridge/internal/health"
	"tcp-proxy-bridge/internal/metrics"
	"tcp-proxy-bridge/internal/source"
)

// main 应用主入口函数
func main() {
	// 初始化日志
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting TCP Proxy Bridge...")

	// 1. 加载配置文件
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
	}
	log.Println("Configuration loaded successfully")

	// 2. 验证源服务器配置
	if err := cfg.ValidateSourceServers(); err != nil {
		log.Fatalf("Source servers configuration validation failed: %v", err)
	}
	log.Println("Source servers configuration validated")

	// 3. 验证身份认证配置
	if err := cfg.ValidateAuthentication(); err != nil {
		log.Fatalf("Authentication configuration validation failed: %v", err)
	}
	log.Println("Authentication configuration validated")

	// 4. 验证心跳配置
	if err := cfg.ValidateHeartbeat(); err != nil {
		log.Fatalf("Heartbeat configuration validation failed: %v", err)
	}
	log.Println("Heartbeat configuration validated")

	// 5. 验证分隔符配置
	if err := cfg.ValidateDelimiter(); err != nil {
		log.Fatalf("Delimiter configuration validation failed: %v", err)
	}
	log.Println("Delimiter configuration validated")

	// 6. 验证目标服务器配置
	if err := cfg.ValidateTargetServers(); err != nil {
		log.Fatalf("Target servers configuration validation failed: %v", err)
	}
	log.Printf("Target servers configuration validated: %d servers configured", len(cfg.TargetServers))

	// 4. 初始化数据库连接
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		db.Close()
		log.Println("Database connection closed")
	}()
	log.Println("Database connection established")

	// 5. 同步目标服务器配置到数据库
	synchronizer := database.NewTargetSynchronizer(db)
	if err := synchronizer.SyncTargetServers(cfg.TargetServers); err != nil {
		log.Fatalf("Failed to sync target servers to database: %v", err)
	}
	log.Println("Target servers synchronized to database")

	// 6. 初始化指标系统
	metrics.Reset()
	log.Println("Metrics system initialized")

	// 7. 从数据库获取启用的目标服务器配置
	targets, err := synchronizer.GetEnabledTargetServers()
	if err != nil {
		log.Fatalf("Failed to get enabled target servers from database: %v", err)
	}
	log.Printf("Loaded %d enabled target servers from database", len(targets))

	// 8. 创建服务实例
	// tcpServer := tcp.NewServer(&cfg.Server, db) // 暂时注释，使用主动连接模式
	forwarderManager := forwarder.NewManager(&cfg.Forwarder, db, targets)
	sourceManager := source.NewManager(cfg) // 传递完整配置
	healthServer := health.NewMinimalServer(cfg.Server.HealthCheckPort, cfg.Server.TCPListenPort, db.DB())

	// 9. 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 10. 启动健康检查服务器
	go func() {
		log.Printf("Starting health server on port %d", cfg.Server.HealthCheckPort)
		if err := healthServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Health server failed: %v", err)
		}
	}()

	// 11. 启动源服务器管理器（主动连接模式）
	go func() {
		log.Println("Starting source server manager in active connection mode...")
		if err := sourceManager.Start(ctx); err != nil {
			log.Fatalf("Source server manager failed to start: %v", err)
		}

		// 设置数据处理回调，将数据存入数据库
		sourceManager.SetDataHandler(func(data []byte) error {
			// 创建消息记录
			message := &database.Message{
				OriginalData: data,
				CreatedAt:    time.Now(),
				Status:       database.StatusPending,
			}

			// 存入数据库
			if err := db.SaveMessage(message); err != nil {
				log.Printf("Failed to save message to database: %v", err)
				return err
			}

			log.Printf("Saved message to database: ID=%d, Size=%d bytes", message.ID, len(data))
			return nil
		})

		// 开始主动连接源服务器
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := sourceManager.ConnectToSource(ctx, func(data []byte) error {
						// 数据处理回调已在上面设置
						return nil
					}); err != nil {
						log.Printf("Source connection error: %v", err)
						time.Sleep(5 * time.Second) // 等待5秒后重试
					}
				}
			}
		}()
	}()

	// 12. 启动转发器管理器
	go func() {
		log.Println("Starting forwarder manager...")
		if err := forwarderManager.Start(ctx); err != nil {
			log.Fatalf("Forwarder manager failed to start: %v", err)
		}
	}()

	// 13. 启动TCP服务器（暂时注释，使用主动连接模式）
	// go func() {
	// 	log.Printf("Starting TCP server on port %d", cfg.Server.TCPListenPort)
	// 	if err := tcpServer.Start(ctx); err != nil {
	// 		log.Fatalf("TCP server failed: %v", err)
	// 	}
	// }()
	log.Println("TCP server startup skipped - using active connection mode to source servers")

	// 14. 启动指标日志记录
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metrics.LogMetrics()
			}
		}
	}()

	// 15. 等待服务启动完成
	time.Sleep(2 * time.Second)

	// 检查初始健康状态
	if healthy, err := healthServer.GetHealthStatus(); !healthy {
		log.Printf("Warning: Service started but health check failed: %v", err)
	} else {
		log.Println("All services started successfully and health check passed")
	}

	log.Println("TCP Proxy Bridge is now running")

	// 16. 设置信号处理，实现优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待终止信号
	sig := <-sigChan
	log.Printf("Received signal: %v, initiating shutdown...", sig)

	// 17. 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 记录关闭前的指标
	metrics.LogMetrics()

	// 停止TCP服务器（暂时注释，使用主动连接模式）
	// log.Println("Stopping TCP server...")
	// tcpServer.Stop(shutdownCtx)
	log.Println("TCP server stop skipped - using active connection mode")

	// 停止转发器管理器
	log.Println("Stopping forwarder manager...")
	forwarderManager.Stop(shutdownCtx)

	// 停止源服务器管理器
	log.Println("Stopping source server manager...")
	sourceManager.Stop(shutdownCtx)

	// 停止健康检查服务器
	log.Println("Stopping health server...")
	healthServer.Stop(shutdownCtx)

	log.Println("TCP Proxy Bridge shutdown completed successfully")
}
