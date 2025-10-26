// internal/source/protocol.go
package source

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

// ProtocolHandler 协议处理器
// 负责处理数据包的解析、防粘包和心跳检测
type ProtocolHandler struct {
	heartbeatInterval time.Duration // 心跳间隔
	lastHeartbeat     time.Time     // 最后心跳时间
	buffer            []byte        // 数据缓冲区
}

// NewProtocolHandler 创建协议处理器
// 参数: heartbeatInterval - 心跳间隔
// 返回: 协议处理器实例
func NewProtocolHandler(heartbeatInterval time.Duration) *ProtocolHandler {
	return &ProtocolHandler{
		heartbeatInterval: heartbeatInterval,
		lastHeartbeat:     time.Now(),
		buffer:            make([]byte, 0, 4096),
	}
}

// HeartbeatPacket 心跳包常量
var HeartbeatPacket = []byte{0xE5, 0xBF, 0x83, 0xE8, 0xB7, 0xB3} // "心跳"的UTF-8编码

// BasePackage 基础数据包结构
type BasePackage struct {
	SourceInfo              uint32    // 信源 (4字节)
	HostInfo                uint32    // 信宿 (4字节)
	PackageNo               uint64    // 包序号 (8字节)
	CurrentDataItem         uint16    // 当前数据项 (2字节)
	DataSumLength           uint32    // 当前数据段长度 (4字节)
	RetransmissionFlag      uint16    // 重复标志 (2字节)
	RetransmissionData      uint16    // 重发数据项 (2字节)
	RetransmissionSumLength uint32    // 重发数据段长度 (4字节)
	Data                    []byte    // 数据内容
	Timestamp               time.Time // 接收时间戳
}

// ProcessData 处理接收到的数据，防粘包
// 参数: data - 接收到的原始数据
// 返回: 完整的数据包列表和错误信息
func (ph *ProtocolHandler) ProcessData(data []byte) ([]*BasePackage, error) {
	// 将新数据添加到缓冲区
	ph.buffer = append(ph.buffer, data...)

	var packages []*BasePackage

	// 循环处理缓冲区中的数据
	for {
		// 检查是否有足够的数据进行解析
		if len(ph.buffer) < 32 { // 最小包头长度
			break
		}

		// 检查是否是心跳包
		if ph.isHeartbeatPacket(ph.buffer) {
			ph.handleHeartbeat()
			ph.buffer = ph.buffer[6:] // 移除心跳包
			continue
		}

		// 解析数据包长度
		dataLength, err := ph.parseDataLength(ph.buffer)
		if err != nil {
			log.Printf("Failed to parse data length: %v", err)
			// 如果解析失败，清空缓冲区
			ph.buffer = ph.buffer[:0]
			break
		}

		// 计算完整数据包长度
		totalLength := 32 + dataLength // 32字节包头 + 数据长度

		// 检查是否有完整的数据包
		if len(ph.buffer) < int(totalLength) {
			break
		}

		// 解析完整数据包
		pkg, err := ph.parsePackage(ph.buffer[:int(totalLength)])
		if err != nil {
			log.Printf("Failed to parse package: %v", err)
			// 移除错误的数据包
			ph.buffer = ph.buffer[int(totalLength):]
			continue
		}

		packages = append(packages, pkg)

		// 从缓冲区移除已处理的数据
		ph.buffer = ph.buffer[int(totalLength):]
	}

	return packages, nil
}

// isHeartbeatPacket 检查是否是心跳包
// 参数: data - 数据
// 返回: 是否是心跳包
func (ph *ProtocolHandler) isHeartbeatPacket(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	return bytes.Equal(data[:6], HeartbeatPacket)
}

// handleHeartbeat 处理心跳包
func (ph *ProtocolHandler) handleHeartbeat() {
	ph.lastHeartbeat = time.Now()
	log.Printf("Received heartbeat packet at %v", ph.lastHeartbeat)
}

// parseDataLength 解析数据长度
// 参数: data - 数据
// 返回: 数据长度和错误信息
func (ph *ProtocolHandler) parseDataLength(data []byte) (uint32, error) {
	if len(data) < 32 {
		return 0, fmt.Errorf("insufficient data for header")
	}

	// 当前数据段长度在第20-23字节 (跳过信源4字节+信宿4字节+包序号8字节+当前数据项2字节+当前数据段长度4字节)
	dataLength := binary.BigEndian.Uint32(data[18:22])

	return dataLength, nil
}

// parsePackage 解析完整数据包
// 参数: data - 完整数据包
// 返回: 解析后的数据包和错误信息
func (ph *ProtocolHandler) parsePackage(data []byte) (*BasePackage, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("insufficient data for package header")
	}

	pkg := &BasePackage{
		Timestamp: time.Now(),
	}

	// 解析包头
	pkg.SourceInfo = binary.BigEndian.Uint32(data[0:4])                // 信源
	pkg.HostInfo = binary.BigEndian.Uint32(data[4:8])                  // 信宿
	pkg.PackageNo = binary.BigEndian.Uint64(data[8:16])                // 包序号
	pkg.CurrentDataItem = binary.BigEndian.Uint16(data[16:18])         // 当前数据项
	pkg.DataSumLength = binary.BigEndian.Uint32(data[18:22])           // 当前数据段长度
	pkg.RetransmissionFlag = binary.BigEndian.Uint16(data[22:24])      // 重复标志
	pkg.RetransmissionData = binary.BigEndian.Uint16(data[24:26])      // 重发数据项
	pkg.RetransmissionSumLength = binary.BigEndian.Uint32(data[26:30]) // 重发数据段长度

	// 解析数据内容
	if len(data) > 32 {
		pkg.Data = make([]byte, len(data)-32)
		copy(pkg.Data, data[32:])
	}

	log.Printf("Parsed package: Source=%d, Host=%d, PackageNo=%d, DataLength=%d",
		pkg.SourceInfo, pkg.HostInfo, pkg.PackageNo, pkg.DataSumLength)

	return pkg, nil
}

// ShouldSendHeartbeat 检查是否应该发送心跳
// 返回: 是否应该发送心跳
func (ph *ProtocolHandler) ShouldSendHeartbeat() bool {
	return time.Since(ph.lastHeartbeat) >= ph.heartbeatInterval
}

// GetHeartbeatPacket 获取心跳包
// 返回: 心跳包数据
func (ph *ProtocolHandler) GetHeartbeatPacket() []byte {
	return HeartbeatPacket
}

// GetLastHeartbeat 获取最后心跳时间
// 返回: 最后心跳时间
func (ph *ProtocolHandler) GetLastHeartbeat() time.Time {
	return ph.lastHeartbeat
}

// ResetHeartbeat 重置心跳时间
func (ph *ProtocolHandler) ResetHeartbeat() {
	ph.lastHeartbeat = time.Now()
}

// GetBufferSize 获取缓冲区大小
// 返回: 缓冲区大小
func (ph *ProtocolHandler) GetBufferSize() int {
	return len(ph.buffer)
}

// ClearBuffer 清空缓冲区
func (ph *ProtocolHandler) ClearBuffer() {
	ph.buffer = ph.buffer[:0]
}

// ToHexString 将字节数组转换为十六进制字符串
// 参数: data - 字节数组
// 返回: 十六进制字符串
func ToHexString(data []byte) string {
	return hex.EncodeToString(data)
}

// FromHexString 将十六进制字符串转换为字节数组
// 参数: hexStr - 十六进制字符串
// 返回: 字节数组和错误信息
func FromHexString(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}
