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
	"tcp-proxy-bridge/internal/tcp"
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

	// 2. 验证目标服务器配置
	if err := cfg.ValidateTargetServers(); err != nil {
		log.Fatalf("Target servers configuration validation failed: %v", err)
	}
	log.Printf("Target servers configuration validated: %d servers configured", len(cfg.TargetServers))

	// 3. 初始化数据库连接
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		db.Close()
		log.Println("Database connection closed")
	}()
	log.Println("Database connection established")

	// 4. 同步目标服务器配置到数据库
	synchronizer := database.NewTargetSynchronizer(db)
	if err := synchronizer.SyncTargetServers(cfg.TargetServers); err != nil {
		log.Fatalf("Failed to sync target servers to database: %v", err)
	}
	log.Println("Target servers synchronized to database")

	// 5. 初始化指标系统
	metrics.Reset()
	log.Println("Metrics system initialized")

	// 6. 从数据库获取启用的目标服务器配置
	targets, err := synchronizer.GetEnabledTargetServers()
	if err != nil {
		log.Fatalf("Failed to get enabled target servers from database: %v", err)
	}
	log.Printf("Loaded %d enabled target servers from database", len(targets))

	// 7. 创建服务实例
	tcpServer := tcp.NewServer(&cfg.Server, db)
	forwarderManager := forwarder.NewManager(&cfg.Forwarder, db, targets)
	healthServer := health.NewMinimalServer(cfg.Server.HealthCheckPort, cfg.Server.TCPListenPort, db.DB())

	// 8. 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 9. 启动健康检查服务器
	go func() {
		log.Printf("Starting health server on port %d", cfg.Server.HealthCheckPort)
		if err := healthServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Health server failed: %v", err)
		}
	}()

	// 10. 启动转发器管理器
	go func() {
		log.Println("Starting forwarder manager...")
		if err := forwarderManager.Start(ctx); err != nil {
			log.Fatalf("Forwarder manager failed to start: %v", err)
		}
	}()

	// 11. 启动TCP服务器
	go func() {
		log.Printf("Starting TCP server on port %d", cfg.Server.TCPListenPort)
		if err := tcpServer.Start(ctx); err != nil {
			log.Fatalf("TCP server failed: %v", err)
		}
	}()

	// 12. 启动指标日志记录
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

	// 13. 等待服务启动完成
	time.Sleep(2 * time.Second)

	// 检查初始健康状态
	if healthy, err := healthServer.GetHealthStatus(); !healthy {
		log.Printf("Warning: Service started but health check failed: %v", err)
	} else {
		log.Println("All services started successfully and health check passed")
	}

	log.Println("TCP Proxy Bridge is now running")

	// 14. 设置信号处理，实现优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待终止信号
	sig := <-sigChan
	log.Printf("Received signal: %v, initiating shutdown...", sig)

	// 15. 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 记录关闭前的指标
	metrics.LogMetrics()

	// 停止TCP服务器
	log.Println("Stopping TCP server...")
	tcpServer.Stop(shutdownCtx)

	// 停止转发器管理器
	log.Println("Stopping forwarder manager...")
	forwarderManager.Stop(shutdownCtx)

	// 停止健康检查服务器
	log.Println("Stopping health server...")
	healthServer.Stop(shutdownCtx)

	log.Println("TCP Proxy Bridge shutdown completed successfully")
}
