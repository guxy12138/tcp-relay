// internal/source/delimiter.go
package source

import (
	"bytes"
	"fmt"
	"log"
)

// DelimiterHandler 分隔符处理器
// 负责处理基于分隔符的数据包边界识别
type DelimiterHandler struct {
	delimiter []byte // 分隔符
	buffer    []byte // 数据缓冲区
	maxLength int    // 最大包长度
}

// NewDelimiterHandler 创建分隔符处理器
// 参数: delimiter - 分隔符, maxLength - 最大包长度
// 返回: 分隔符处理器实例
func NewDelimiterHandler(delimiter []byte, maxLength int) *DelimiterHandler {
	return &DelimiterHandler{
		delimiter: delimiter,
		buffer:    make([]byte, 0, 4096),
		maxLength: maxLength,
	}
}

// ProcessData 处理接收到的数据，基于分隔符分割数据包
// 参数: data - 接收到的原始数据
// 返回: 完整的数据包列表和错误信息
func (dh *DelimiterHandler) ProcessData(data []byte) ([][]byte, error) {
	// 将新数据添加到缓冲区
	dh.buffer = append(dh.buffer, data...)

	var packets [][]byte

	// 循环处理缓冲区中的数据
	for {
		// 查找分隔符位置
		delimiterIndex := bytes.Index(dh.buffer, dh.delimiter)

		if delimiterIndex == -1 {
			// 没有找到分隔符
			// 检查缓冲区是否超过最大长度
			if len(dh.buffer) > dh.maxLength {
				log.Printf("Buffer overflow: %d bytes, max allowed: %d", len(dh.buffer), dh.maxLength)
				// 清空缓冲区，丢弃数据
				dh.buffer = dh.buffer[:0]
			}
			break
		}

		// 提取完整数据包（不包含分隔符）
		packet := make([]byte, delimiterIndex)
		copy(packet, dh.buffer[:delimiterIndex])

		// 检查包是否为空
		if len(packet) > 0 {
			packets = append(packets, packet)
			log.Printf("Extracted packet: %d bytes", len(packet))
		}

		// 从缓冲区移除已处理的数据（包括分隔符）
		dh.buffer = dh.buffer[delimiterIndex+len(dh.delimiter):]
	}

	return packets, nil
}

// GetBufferSize 获取缓冲区大小
// 返回: 缓冲区大小
func (dh *DelimiterHandler) GetBufferSize() int {
	return len(dh.buffer)
}

// ClearBuffer 清空缓冲区
func (dh *DelimiterHandler) ClearBuffer() {
	dh.buffer = dh.buffer[:0]
}

// GetDelimiter 获取分隔符
// 返回: 分隔符字节数组
func (dh *DelimiterHandler) GetDelimiter() []byte {
	return dh.delimiter
}

// SetMaxLength 设置最大包长度
// 参数: maxLength - 最大包长度
func (dh *DelimiterHandler) SetMaxLength(maxLength int) {
	dh.maxLength = maxLength
}

// GetMaxLength 获取最大包长度
// 返回: 最大包长度
func (dh *DelimiterHandler) GetMaxLength() int {
	return dh.maxLength
}

// AddDelimiter 为数据包添加分隔符
// 参数: data - 数据包
// 返回: 添加分隔符后的数据
func (dh *DelimiterHandler) AddDelimiter(data []byte) []byte {
	result := make([]byte, len(data)+len(dh.delimiter))
	copy(result, data)
	copy(result[len(data):], dh.delimiter)
	return result
}

// RemoveDelimiter 从数据包中移除分隔符
// 参数: data - 包含分隔符的数据
// 返回: 移除分隔符后的数据和是否包含分隔符
func (dh *DelimiterHandler) RemoveDelimiter(data []byte) ([]byte, bool) {
	if bytes.HasSuffix(data, dh.delimiter) {
		return data[:len(data)-len(dh.delimiter)], true
	}
	return data, false
}

// ValidatePacket 验证数据包格式
// 参数: packet - 数据包
// 返回: 是否有效
func (dh *DelimiterHandler) ValidatePacket(packet []byte) bool {
	// 检查包长度
	if len(packet) == 0 {
		return false
	}

	// 检查是否超过最大长度
	if len(packet) > dh.maxLength {
		return false
	}

	// 检查是否包含分隔符（包内不应该包含分隔符）
	if bytes.Contains(packet, dh.delimiter) {
		return false
	}

	return true
}

// GetStatus 获取处理器状态
// 返回: 状态信息
func (dh *DelimiterHandler) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"delimiter":    fmt.Sprintf("%x", dh.delimiter),
		"buffer_size":  len(dh.buffer),
		"max_length":   dh.maxLength,
		"buffer_usage": float64(len(dh.buffer)) / float64(dh.maxLength) * 100,
	}
}
