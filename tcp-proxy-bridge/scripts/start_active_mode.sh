#!/bin/bash

# scripts/start_active_mode.sh
# 启动主动连接模式的TCP代理桥接服务

set -e

# 设置配置文件路径
CONFIG_FILE=${CONFIG_FILE:-"configs/config_active_mode.yaml"}

# 检查配置文件是否存在
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件 $CONFIG_FILE 不存在"
    echo "请确保配置文件存在，或设置 CONFIG_FILE 环境变量"
    exit 1
fi

# 检查可执行文件是否存在
BINARY_PATH="./bin/tcp-proxy-bridge"
if [ ! -f "$BINARY_PATH" ]; then
    echo "构建可执行文件..."
    go build -o "$BINARY_PATH" ./cmd/server
    if [ $? -ne 0 ]; then
        echo "错误: 构建失败"
        exit 1
    fi
fi

echo "启动TCP代理桥接服务 (主动连接模式)..."
echo "配置文件: $CONFIG_FILE"
echo "可执行文件: $BINARY_PATH"
echo ""

# 设置环境变量并启动服务
export CONFIG_FILE="$CONFIG_FILE"
exec "$BINARY_PATH"
