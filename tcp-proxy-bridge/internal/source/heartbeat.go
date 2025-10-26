// internal/source/heartbeat.go
package source

import (
	"bytes"
	"log"
	"time"
)

// HeartbeatManager 心跳管理器
// 负责处理与源服务器的心跳机制
type HeartbeatManager struct {
	interval           time.Duration // 心跳间隔
	lastHeartbeat      time.Time     // 最后心跳时间
	lastHeartbeatRecv  time.Time     // 最后接收心跳时间
	writeIdleTimeout   time.Duration // 写空闲超时
	readIdleTimeout    time.Duration // 读空闲超时
	separatorCharacter []byte        // 分隔符
}

// NewHeartbeatManager 创建心跳管理器
// 参数: interval - 心跳间隔, writeIdle - 写空闲超时, readIdle - 读空闲超时
// 返回: 心跳管理器实例
func NewHeartbeatManager(interval, writeIdle, readIdle time.Duration) *HeartbeatManager {
	return &HeartbeatManager{
		interval:           interval,
		writeIdleTimeout:   writeIdle,
		readIdleTimeout:    readIdle,
		separatorCharacter: []byte{0x78, 0x78, 0x78, 0x78, 0x88, 0x88, 0x88, 0x88}, // "7878787888888888"
	}
}

// GenerateHeartbeatPacket 生成心跳包 (XFType203)
// 返回: 心跳包数据
func (hm *HeartbeatManager) GenerateHeartbeatPacket() []byte {
	// 心跳包内容: "心跳"的UTF-8编码 + 分隔符
	heartbeatContent := []byte{0xE5, 0xBF, 0x83, 0xE8, 0xB7, 0xB3} // "心跳"
	
	// 组合心跳内容和分隔符
	packet := make([]byte, len(heartbeatContent)+len(hm.separatorCharacter))
	copy(packet, heartbeatContent)
	copy(packet[len(heartbeatContent):], hm.separatorCharacter)
	
	hm.lastHeartbeat = time.Now()
	
	log.Printf("Generated heartbeat packet: %d bytes", len(packet))
	return packet
}

// ShouldSendHeartbeat 检查是否应该发送心跳
// 返回: 是否应该发送心跳
func (hm *HeartbeatManager) ShouldSendHeartbeat() bool {
	return time.Since(hm.lastHeartbeat) >= hm.interval
}

// IsWriteIdle 检查是否写空闲超时
// 返回: 是否写空闲超时
func (hm *HeartbeatManager) IsWriteIdle() bool {
	return time.Since(hm.lastHeartbeat) >= hm.writeIdleTimeout
}

// IsReadIdle 检查是否读空闲超时
// 返回: 是否读空闲超时
func (hm *HeartbeatManager) IsReadIdle() bool {
	return time.Since(hm.lastHeartbeatRecv) >= hm.readIdleTimeout
}

// UpdateHeartbeatSent 更新心跳发送时间
func (hm *HeartbeatManager) UpdateHeartbeatSent() {
	hm.lastHeartbeat = time.Now()
}

// UpdateHeartbeatReceived 更新心跳接收时间
func (hm *HeartbeatManager) UpdateHeartbeatReceived() {
	hm.lastHeartbeatRecv = time.Now()
}

// IsHeartbeatPacket 检查是否是心跳包
// 参数: data - 数据
// 返回: 是否是心跳包
func (hm *HeartbeatManager) IsHeartbeatPacket(data []byte) bool {
	// 检查是否包含心跳内容
	heartbeatContent := []byte{0xE5, 0xBF, 0x83, 0xE8, 0xB7, 0xB3}
	return bytes.Contains(data, heartbeatContent)
}

// IsSeparatorPacket 检查是否是分隔符包
// 参数: data - 数据
// 返回: 是否是分隔符包
func (hm *HeartbeatManager) IsSeparatorPacket(data []byte) bool {
	return bytes.Contains(data, hm.separatorCharacter)
}

// GetSeparatorCharacter 获取分隔符
// 返回: 分隔符字节数组
func (hm *HeartbeatManager) GetSeparatorCharacter() []byte {
	return hm.separatorCharacter
}

// GetLastHeartbeatTime 获取最后心跳时间
// 返回: 最后心跳时间
func (hm *HeartbeatManager) GetLastHeartbeatTime() time.Time {
	return hm.lastHeartbeat
}

// GetLastHeartbeatRecvTime 获取最后接收心跳时间
// 返回: 最后接收心跳时间
func (hm *HeartbeatManager) GetLastHeartbeatRecvTime() time.Time {
	return hm.lastHeartbeatRecv
}

// Reset 重置心跳管理器
func (hm *HeartbeatManager) Reset() {
	hm.lastHeartbeat = time.Now()
	hm.lastHeartbeatRecv = time.Now()
}

// GetStatus 获取心跳状态
// 返回: 状态信息
func (hm *HeartbeatManager) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"interval":              hm.interval.String(),
		"write_idle_timeout":    hm.writeIdleTimeout.String(),
		"read_idle_timeout":     hm.readIdleTimeout.String(),
		"last_heartbeat":        hm.lastHeartbeat,
		"last_heartbeat_recv":   hm.lastHeartbeatRecv,
		"should_send_heartbeat": hm.ShouldSendHeartbeat(),
		"is_write_idle":         hm.IsWriteIdle(),
		"is_read_idle":          hm.IsReadIdle(),
	}
}
