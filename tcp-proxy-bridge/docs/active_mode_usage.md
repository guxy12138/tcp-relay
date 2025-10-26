# 主动连接模式使用说明

## 概述

TCP代理桥接服务现在支持**主动连接模式**，即主动连接到源服务器获取数据，而不是被动接收连接。这种模式更适合从外部数据源主动拉取数据的场景。

## 架构变化

### 之前（被动模式）
```
外部客户端 → TCP服务器(9999端口) → 数据库 → 转发器 → 目标服务器
```

### 现在（主动模式）
```
源服务器 ← 主动连接 ← 源服务器管理器 → 数据库 → 转发器 → 目标服务器
```

## 主要特性

1. **主动连接**：主动连接到配置的源服务器
2. **主备切换**：支持主备服务器自动切换
3. **防粘包处理**：自动处理数据包边界
4. **心跳机制**：60秒心跳保持连接
5. **数据存储**：接收的数据自动存入数据库
6. **转发功能**：转发器继续工作，将数据发送到目标服务器

## 配置说明

### 1. 源服务器配置

在 `configs/config_active_mode.yaml` 中配置源服务器：

```yaml
source_servers:
  primary:
    id: "primary-server"
    name: "主源服务器"
    address: "192.168.1.100:8888"  # 主服务器地址
    enabled: true
    timeout: "10s"
    max_retries: 5
    failover_threshold: 3          # 连续失败3次后切换
    
  backup:
    id: "backup-server"
    name: "备用源服务器"
    address: "192.168.1.101:8888"  # 备用服务器地址
    enabled: true
    timeout: "10s"
    max_retries: 3
    failover_threshold: 3
```

### 2. 目标服务器配置

目标服务器配置保持不变：

```yaml
target_servers:
  - id: "target-1"
    name: "目标服务器1"
    address: "192.168.1.200:9999"
    enabled: true
    priority: 1
```

## 使用方法

### 1. 快速启动

使用提供的启动脚本：

```bash
# 使用默认配置文件
./scripts/start_active_mode.sh

# 或指定配置文件
CONFIG_FILE=configs/config_active_mode.yaml ./scripts/start_active_mode.sh
```

### 2. 手动启动

```bash
# 构建
go build -o ./bin/tcp-proxy-bridge ./cmd/server

# 启动
CONFIG_FILE=configs/config_active_mode.yaml ./bin/tcp-proxy-bridge
```

### 3. Docker启动

```bash
# 构建镜像
docker build -t tcp-proxy-bridge .

# 运行容器
docker run -d \
  --name tcp-proxy-bridge \
  -v $(pwd)/configs/config_active_mode.yaml:/app/config.yaml \
  tcp-proxy-bridge
```

## 数据流程

### 1. 数据接收流程

```
源服务器 → 主动连接 → 协议解析 → 防粘包处理 → 数据库存储
```

### 2. 数据转发流程

```
数据库 → 转发器 → 目标服务器1
       → 转发器 → 目标服务器2
       → 转发器 → 目标服务器3
```

### 3. 故障切换流程

```
主服务器故障 → 检测失败 → 切换备用服务器 → 继续接收数据
主服务器恢复 → 健康检查 → 切换回主服务器
```

## 监控和日志

### 1. 关键日志信息

- **连接状态**：`Starting source server manager in active connection mode`
- **数据接收**：`Saved message to database: ID=123, Size=1024 bytes`
- **故障切换**：`Performing failover: 主源服务器 -> 备用源服务器`
- **心跳状态**：`Sent heartbeat packet to 主源服务器`

### 2. 健康检查

访问健康检查端点：

```bash
curl http://localhost:8080/health
```

### 3. 状态监控

通过日志查看当前状态：

```bash
# 查看连接状态
tail -f logs/tcp-proxy-bridge.log | grep "Source connection"

# 查看数据接收
tail -f logs/tcp-proxy-bridge.log | grep "Saved message"

# 查看故障切换
tail -f logs/tcp-proxy-bridge.log | grep "failover"
```

## 配置参数说明

### 源服务器参数

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `address` | 服务器地址 | - | `192.168.1.100:8888` |
| `timeout` | 连接超时 | `10s` | `10s`, `30s` |
| `max_retries` | 最大重试次数 | `3` | `3`, `5` |
| `failover_threshold` | 故障切换阈值 | `3` | `3`, `5` |
| `health_check_interval` | 健康检查间隔 | `30s` | `30s`, `60s` |

### 心跳参数

- **心跳间隔**：60秒
- **心跳包内容**：`E5BF83E8B7B3` (UTF-8编码的"心跳")
- **心跳超时**：5秒

## 故障排查

### 1. 连接问题

**问题**：无法连接到源服务器
**排查**：
```bash
# 检查网络连通性
telnet 192.168.1.100 8888

# 检查防火墙
sudo ufw status

# 查看连接日志
grep "Connection error" logs/tcp-proxy-bridge.log
```

### 2. 数据接收问题

**问题**：没有接收到数据
**排查**：
```bash
# 检查源服务器是否发送数据
# 检查协议解析是否正确
# 查看数据库是否有新记录
```

### 3. 故障切换问题

**问题**：故障切换不工作
**排查**：
```bash
# 检查故障阈值配置
# 查看健康检查日志
# 确认备用服务器配置正确
```

## 性能优化

### 1. 连接池配置

```yaml
database:
  max_open_conns: 50      # 增加连接池大小
  max_idle_conns: 25      # 增加空闲连接数
  conn_max_lifetime: "5m" # 连接生命周期
```

### 2. 批处理配置

```yaml
forwarder:
  batch_size: 100              # 增加批处理大小
  max_processing_workers: 10   # 增加工作线程数
  process_interval: "5s"       # 调整处理间隔
```

### 3. 缓冲区配置

```yaml
server:
  max_message_size: 65536      # 增加消息大小限制
  read_timeout: 30s           # 调整读取超时
  write_timeout: 30s          # 调整写入超时
```

## 注意事项

1. **网络稳定性**：确保到源服务器的网络连接稳定
2. **服务器配置**：确保源服务器支持TCP连接
3. **数据格式**：确保数据格式符合协议规范
4. **资源监控**：监控内存和CPU使用情况
5. **日志管理**：定期清理日志文件

## 回退到被动模式

如果需要回退到被动模式（TCP服务器模式），只需：

1. 取消注释 `cmd/server/main.go` 中的TCP服务器相关代码
2. 注释掉源服务器管理器的主动连接代码
3. 使用原来的配置文件

这样可以在两种模式之间灵活切换。
