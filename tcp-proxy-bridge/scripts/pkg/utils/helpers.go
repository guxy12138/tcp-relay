// pkg/utils/helpers.go
package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// GenerateMessageID 生成消息唯一标识
// 参数: sourceIP - 源IP地址, data - 消息数据
// 返回: 消息ID字符串
func GenerateMessageID(sourceIP string, data []byte) string {
	timestamp := time.Now().UnixNano()
	// 使用SHA256生成数据哈希
	hash := sha256.Sum256(append([]byte(sourceIP), data...))
	return fmt.Sprintf("%d-%s", timestamp, hex.EncodeToString(hash[:8]))
}

// IsValidIP 验证IP地址格式是否有效
// 参数: ip - IP地址字符串
// 返回: 是否有效
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidPort 验证端口号是否有效
// 参数: port - 端口号
// 返回: 是否有效
func IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}

// ParseAddress 解析地址字符串为IP和端口
// 参数: addr - 地址字符串 (格式: "ip:port")
// 返回: IP地址, 端口号, 错误信息
func ParseAddress(addr string) (string, int, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid address format: %s, expected 'ip:port'", addr)
	}

	ip := parts[0]
	if !IsValidIP(ip) {
		return "", 0, fmt.Errorf("invalid IP address: %s", ip)
	}

	var port int
	if _, err := fmt.Sscanf(parts[1], "%d", &port); err != nil {
		return "", 0, fmt.Errorf("invalid port: %s", parts[1])
	}

	if !IsValidPort(port) {
		return "", 0, fmt.Errorf("port out of range: %d", port)
	}

	return ip, port, nil
}

// RetryWithBackoff 带指数退避的重试机制
// 参数: maxRetries - 最大重试次数, baseDelay - 基础延迟, fn - 要重试的函数
// 返回: 错误信息
func RetryWithBackoff(maxRetries int, baseDelay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil // 成功，返回
		}

		if i == maxRetries-1 {
			break // 最后一次尝试失败，跳出循环
		}

		// 指数退避延迟
		delay := baseDelay * time.Duration(1<<uint(i))
		log.Printf("Attempt %d failed, retrying in %v: %v", i+1, delay, err)
		time.Sleep(delay)
	}
	return fmt.Errorf("after %d attempts, last error: %v", maxRetries, err)
}

// FormatBytes 格式化字节大小为人类可读的格式
// 参数: bytes - 字节数
// 返回: 格式化后的字符串
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetLocalIP 获取本地IP地址
// 返回: 本地IP地址字符串
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}
