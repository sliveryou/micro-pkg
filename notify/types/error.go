package types

import (
	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/internal/bizerr"
)

var (
	// ErrInvalidParams 请求参数错误
	ErrInvalidParams = errcode.ErrInvalidParams

	// ErrProviderOverQuota 提供方请求超过配额错误
	ErrProviderOverQuota = bizerr.ErrProviderOverQuota
	// ErrIPSourceOverQuota IP 地址请求超过配额错误
	ErrIPSourceOverQuota = bizerr.ErrIPSourceOverQuota
	// ErrReceiverOverQuota 接收方请求超过配额错误
	ErrReceiverOverQuota = bizerr.ErrReceiverOverQuota

	// ErrSendTooFrequently 发送过于频繁错误
	ErrSendTooFrequently = bizerr.ErrSendTooFrequently
	// ErrVerifyTooFrequently 验证过于频繁错误
	ErrVerifyTooFrequently = bizerr.ErrVerifyTooFrequently

	// ErrEmailUnsupported 暂不支持邮件通知服务错误
	ErrEmailUnsupported = bizerr.ErrEmailUnsupported
	// ErrSmsUnsupported 暂不支持短信通知服务错误
	ErrSmsUnsupported = bizerr.ErrSmsUnsupported
	// ErrEmailTmplNotFound 邮件模板不存在错误
	ErrEmailTmplNotFound = bizerr.ErrEmailTmplNotFound
	// ErrSmsTmplNotFound 短信模板不存在错误
	ErrSmsTmplNotFound = bizerr.ErrSmsTmplNotFound

	// ErrInvalidCaptcha 验证码错误
	ErrInvalidCaptcha = bizerr.ErrInvalidCaptcha
	// ErrCaptchaNotFound 验证码不存在或已过期错误
	ErrCaptchaNotFound = bizerr.ErrCaptchaNotFound
)
