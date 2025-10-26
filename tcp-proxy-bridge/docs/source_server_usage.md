# 源服务器主备切换功能使用说明

## 功能概述

本系统实现了源服务器的主备切换功能，支持：
- 主备服务器自动切换
- 防粘包处理
- 心跳机制（60秒间隔）
- 健康检查
- 故障检测和恢复

## 配置说明

### 配置文件结构

在 `configs/config.yaml` 中添加源服务器配置：

```yaml
# 源服务器配置 - 支持主备切换
source_servers:
  primary:
    id: "primary-server"              # 主服务器唯一标识
    name: "主源服务器"                # 服务器显示名称
    address: "192.168.1.100:8888"     # 服务器地址
    enabled: true                     # 是否启用
    timeout: "10s"                    # 连接超时时间
    max_retries: 5                    # 最大重试次数
    batch_size: 100                   # 批量处理大小
    health_check_interval: "30s"      # 健康检查间隔
    health_check_timeout: "5s"        # 健康检查超时
    failover_threshold: 3             # 故障切换阈值（连续失败次数）
    
  backup:
    id: "backup-server"               # 备用服务器唯一标识
    name: "备用源服务器"              # 服务器显示名称
    address: "192.168.1.101:8888"     # 服务器地址
    enabled: true                     # 是否启用
    timeout: "10s"                    # 连接超时时间
    max_retries: 3                    # 最大重试次数
    batch_size: 50                    # 批量处理大小
    health_check_interval: "30s"      # 健康检查间隔
    health_check_timeout: "5s"        # 健康检查超时
    failover_threshold: 3             # 故障切换阈值（连续失败次数）
```

## 数据协议

### 心跳包格式
- 心跳包内容：`E5BF83E8B7B3` (UTF-8编码的"心跳")
- 发送间隔：60秒

### 数据包格式
数据包采用固定包头格式：

```
+--------+--------+--------+--------+--------+--------+--------+--------+
| 信源(4字节) | 信宿(4字节) | 包序号(8字节) | 当前数据项(2字节) | 当前数据段长度(4字节) |
+--------+--------+--------+--------+--------+--------+--------+--------+
| 重复标志(2字节) | 重发数据项(2字节) | 重发数据段长度(4字节) | 数据内容(变长) |
+--------+--------+--------+--------+--------+--------+--------+--------+
```

## 使用方法

### 1. 基本使用

```go
package main

import (
    "context"
    "log"
    "time"
    
    "tcp-proxy-bridge/internal/config"
    "tcp-proxy-bridge/internal/source"
)

func main() {
    // 加载配置
    cfg, err := config.LoadConfig("configs/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 验证源服务器配置
    if err := cfg.ValidateSourceServers(); err != nil {
        log.Fatalf("Source servers validation failed: %v", err)
    }
    
    // 创建源服务器管理器
    manager := source.NewManager(&cfg.SourceServers)
    
    // 设置数据处理回调
    manager.SetDataHandler(func(data []byte) error {
        log.Printf("Received data: %d bytes", len(data))
        // 处理接收到的数据
        return nil
    })
    
    // 启动管理器
    ctx := context.Background()
    if err := manager.Start(ctx); err != nil {
        log.Fatalf("Failed to start manager: %v", err)
    }
    
    // 连接到源服务器
    go func() {
        for {
            if err := manager.ConnectToSource(ctx, func(data []byte) error {
                // 处理数据
                return nil
            }); err != nil {
                log.Printf("Connection error: %v", err)
                time.Sleep(5 * time.Second)
            }
        }
    }()
    
    // 保持运行
    select {}
}
```

### 2. 在主服务器中集成

源服务器管理器已经集成到主服务器中，启动时会自动：

1. 验证源服务器配置
2. 启动源服务器管理器
3. 开始健康检查
4. 建立连接并开始接收数据

## 功能特性

### 1. 主备切换
- 当主服务器连续失败达到阈值时，自动切换到备用服务器
- 当主服务器恢复时，自动切换回主服务器
- 支持手动配置故障切换阈值

### 2. 防粘包处理
- 自动解析数据包头，获取数据长度
- 确保完整数据包的接收
- 处理不完整的数据包

### 3. 心跳机制
- 每60秒发送一次心跳包
- 自动检测心跳响应
- 心跳失败时触发故障切换

### 4. 健康检查
- 定期检查服务器连接状态
- 可配置检查间隔和超时时间
- 支持TCP连接测试

### 5. 故障恢复
- 自动重试机制
- 指数退避算法
- 故障统计和监控

## 监控和调试

### 获取状态信息

```go
status := manager.GetStatus()
log.Printf("Manager status: %+v", status)
```

状态信息包括：
- `is_running`: 管理器是否运行
- `current_server`: 当前使用的服务器
- `failure_counts`: 各服务器失败次数
- `last_fail_time`: 各服务器最后失败时间

### 日志输出

系统会输出详细的日志信息，包括：
- 连接状态变化
- 心跳包发送/接收
- 数据包解析结果
- 故障切换事件
- 健康检查结果

## 注意事项

1. **网络配置**：确保主备服务器网络可达
2. **端口配置**：确保端口未被占用
3. **超时设置**：根据网络环境调整超时时间
4. **故障阈值**：根据业务需求调整故障切换阈值
5. **心跳间隔**：可根据需要调整心跳发送间隔

## 故障排查

### 常见问题

1. **连接失败**
   - 检查服务器地址和端口
   - 确认网络连通性
   - 检查防火墙设置

2. **心跳超时**
   - 检查网络延迟
   - 调整心跳超时时间
   - 确认服务器响应

3. **数据包解析错误**
   - 检查数据格式
   - 确认包头结构
   - 查看详细日志

### 调试建议

1. 启用详细日志输出
2. 监控网络连接状态
3. 检查系统资源使用情况
4. 验证配置参数设置
