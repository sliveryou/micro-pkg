package types

import (
	"time"

	"github.com/sliveryou/go-tool/v2/validator"
)

//go:generate enumer -type NotifyMethod -json -linecomment -output param_string.go

// NotifyMethod 通知方式
type NotifyMethod int32

const (
	// Sms 通知方式：短信
	Sms NotifyMethod = 0 // sms
	// Email 通知方式：邮件
	Email NotifyMethod = 1 // email
)

// SendParams 发送通知参数
type SendParams struct {
	CommonParams         // 通用通知参数
	IsMock       bool    // 是否模拟发送
	Params       []Param // 参数列表
}

// IsValid 判断参数是否合法
func (p *SendParams) IsValid() bool {
	if p == nil {
		return false
	}

	return p.CommonParams.IsValid()
}

// VerifyParams 校验通知参数
type VerifyParams struct {
	CommonParams        // 通用通知参数
	Code         string // 验证码
	Clear        bool   // 验证成功后是否清除
}

// IsValid 判断参数是否合法
func (p *VerifyParams) IsValid() bool {
	if p == nil {
		return false
	}

	return p.CommonParams.IsValid() && p.Code != ""
}

// CommonParams 通用通知参数
type CommonParams struct {
	NotifyMethod NotifyMethod // 通知方式（可以为 Sms 或 Email）
	IP           string       // IP 地址
	Provider     string       // 提供方
	Receiver     string       // 接受方
	TemplateID   string       // 模板编号
}

// IsValid 判断参数是否合法
func (p CommonParams) IsValid() bool {
	if p.IP == "" || p.Provider == "" || p.TemplateID == "" {
		return false
	}

	switch p.NotifyMethod {
	case Sms:
		if err := validator.VerifyVar(p.Receiver, "len=11,number"); err != nil {
			return false
		}
	case Email:
		if err := validator.VerifyVar(p.Receiver, "email"); err != nil {
			return false
		}
	default:
		return false
	}

	return true
}

// CommonParam 通用参数
type CommonParam struct {
	Key   string // 键
	Value string // 值
}

// GetKey 获取参数键
func (p *CommonParam) GetKey() string {
	return p.Key
}

// GetValue 获取参数值
func (p *CommonParam) GetValue() string {
	return p.Value
}

// CodeParam 验证码参数
type CodeParam struct {
	Key        string        // 键
	Value      string        // 值
	Length     int           // 长度
	Expiration time.Duration // 过期时间
}

// GetKey 获取参数键
func (p *CodeParam) GetKey() string {
	return p.Key
}

// GetValue 获取参数值
func (p *CodeParam) GetValue() string {
	return p.Value
}

// IsEmpty 判断是否为空
func (p *CodeParam) IsEmpty() bool {
	return p.Key == "" && p.Value == "" && p.Length == 0
}
