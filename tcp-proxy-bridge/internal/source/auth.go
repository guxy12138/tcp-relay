// internal/source/auth.go
package source

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

// AuthManager 身份认证管理器
// 负责处理与源服务器的身份认证
type AuthManager struct {
	token        string    // 认证令牌
	sourceID     uint32    // 信源ID
	hostID       uint32    // 信宿ID
	packageNo    uint64    // 包序号
	lastAuthTime time.Time // 最后认证时间
}

// NewAuthManager 创建身份认证管理器
// 参数: token - 认证令牌, sourceID - 信源ID, hostID - 信宿ID
// 返回: 身份认证管理器实例
func NewAuthManager(token string, sourceID, hostID uint32) *AuthManager {
	return &AuthManager{
		token:     token,
		sourceID:  sourceID,
		hostID:    hostID,
		packageNo: 1,
	}
}

// GenerateAuthPacket 生成身份认证包 (XFType100)
// 返回: 认证包数据和错误信息
func (am *AuthManager) GenerateAuthPacket() ([]byte, error) {
	// 创建基础包结构
	basePackage := &BasePackage{
		SourceInfo:              am.sourceID,  // 信源 (0x322)
		HostInfo:                am.hostID,    // 信宿 (0x0014)
		PackageNo:               am.packageNo, // 包序号
		CurrentDataItem:         1,            // 当前数据项
		DataSumLength:           0x0028,       // 当前数据段长度 (40字节)
		RetransmissionFlag:      0x00,         // 重复标志
		RetransmissionData:      0x00,         // 重发数据项
		RetransmissionSumLength: 0x0000,       // 重发数据段长度
		Timestamp:               time.Now(),
	}

	// 生成32字节的token数据
	tokenBytes, err := am.generateTokenBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	// 设置数据内容为token
	basePackage.Data = tokenBytes

	// 序列化包
	packetData, err := am.serializeBasePackage(basePackage)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize auth packet: %v", err)
	}

	// 增加包序号
	am.packageNo++
	am.lastAuthTime = time.Now()

	log.Printf("Generated auth packet: SourceID=0x%X, HostID=0x%X, PackageNo=%d, TokenLength=%d",
		am.sourceID, am.hostID, am.packageNo-1, len(tokenBytes))

	return packetData, nil
}

// generateTokenBytes 生成token字节数组
// 返回: token字节数组和错误信息
func (am *AuthManager) generateTokenBytes() ([]byte, error) {
	if am.token == "" {
		// 如果没有提供token，生成随机token
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			return nil, err
		}
		return tokenBytes, nil
	}

	// 如果token是十六进制字符串，转换为字节数组
	if len(am.token) == 64 { // 32字节 = 64个十六进制字符
		return hex.DecodeString(am.token)
	}

	// 如果token是普通字符串，转换为ASCII字节数组
	if len(am.token) <= 32 {
		tokenBytes := make([]byte, 32)
		copy(tokenBytes, []byte(am.token))
		return tokenBytes, nil
	}

	return nil, fmt.Errorf("invalid token format: %s", am.token)
}

// serializeBasePackage 序列化基础包
// 参数: pkg - 基础包
// 返回: 序列化后的字节数组和错误信息
func (am *AuthManager) serializeBasePackage(pkg *BasePackage) ([]byte, error) {
	// 计算总长度：32字节包头 + 数据长度
	totalLength := 32 + len(pkg.Data)
	data := make([]byte, totalLength)

	// 序列化包头 (大端序)
	offset := 0

	// 信源 (4字节)
	data[offset] = byte(pkg.SourceInfo >> 24)
	data[offset+1] = byte(pkg.SourceInfo >> 16)
	data[offset+2] = byte(pkg.SourceInfo >> 8)
	data[offset+3] = byte(pkg.SourceInfo)
	offset += 4

	// 信宿 (4字节)
	data[offset] = byte(pkg.HostInfo >> 24)
	data[offset+1] = byte(pkg.HostInfo >> 16)
	data[offset+2] = byte(pkg.HostInfo >> 8)
	data[offset+3] = byte(pkg.HostInfo)
	offset += 4

	// 包序号 (8字节)
	data[offset] = byte(pkg.PackageNo >> 56)
	data[offset+1] = byte(pkg.PackageNo >> 48)
	data[offset+2] = byte(pkg.PackageNo >> 40)
	data[offset+3] = byte(pkg.PackageNo >> 32)
	data[offset+4] = byte(pkg.PackageNo >> 24)
	data[offset+5] = byte(pkg.PackageNo >> 16)
	data[offset+6] = byte(pkg.PackageNo >> 8)
	data[offset+7] = byte(pkg.PackageNo)
	offset += 8

	// 当前数据项 (2字节)
	data[offset] = byte(pkg.CurrentDataItem >> 8)
	data[offset+1] = byte(pkg.CurrentDataItem)
	offset += 2

	// 当前数据段长度 (4字节)
	data[offset] = byte(pkg.DataSumLength >> 24)
	data[offset+1] = byte(pkg.DataSumLength >> 16)
	data[offset+2] = byte(pkg.DataSumLength >> 8)
	data[offset+3] = byte(pkg.DataSumLength)
	offset += 4

	// 重复标志 (2字节)
	data[offset] = byte(pkg.RetransmissionFlag >> 8)
	data[offset+1] = byte(pkg.RetransmissionFlag)
	offset += 2

	// 重发数据项 (2字节)
	data[offset] = byte(pkg.RetransmissionData >> 8)
	data[offset+1] = byte(pkg.RetransmissionData)
	offset += 2

	// 重发数据段长度 (4字节)
	data[offset] = byte(pkg.RetransmissionSumLength >> 24)
	data[offset+1] = byte(pkg.RetransmissionSumLength >> 16)
	data[offset+2] = byte(pkg.RetransmissionSumLength >> 8)
	data[offset+3] = byte(pkg.RetransmissionSumLength)
	offset += 4

	// 数据内容
	copy(data[offset:], pkg.Data)

	return data, nil
}

// ShouldReauth 检查是否需要重新认证
// 参数: interval - 认证间隔
// 返回: 是否需要重新认证
func (am *AuthManager) ShouldReauth(interval time.Duration) bool {
	return time.Since(am.lastAuthTime) >= interval
}

// GetToken 获取当前token
// 返回: token字符串
func (am *AuthManager) GetToken() string {
	return am.token
}

// GetSourceID 获取信源ID
// 返回: 信源ID
func (am *AuthManager) GetSourceID() uint32 {
	return am.sourceID
}

// GetHostID 获取信宿ID
// 返回: 信宿ID
func (am *AuthManager) GetHostID() uint32 {
	return am.hostID
}

// GetPackageNo 获取当前包序号
// 返回: 包序号
func (am *AuthManager) GetPackageNo() uint64 {
	return am.packageNo
}
