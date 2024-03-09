package types

import "github.com/sliveryou/micro-pkg/errcode"

var (
	// ErrInvalidParams 请求参数错误
	ErrInvalidParams = errcode.ErrInvalidParams

	// ErrProviderOverQuota 提供方请求超过配额错误
	ErrProviderOverQuota = errcode.NewCommon("该提供方今日请求超过配额，请明日再试")
	// ErrIPSourceOverQuota IP 地址请求超过配额错误
	ErrIPSourceOverQuota = errcode.NewCommon("该 IP 地址今日请求超过配额，请明日再试")
	// ErrReceiverOverQuota 接收方请求超过配额错误
	ErrReceiverOverQuota = errcode.NewCommon("该接收方今日请求超过配额，请明日再试")

	// ErrSendTooFrequently 发送过于频繁错误
	ErrSendTooFrequently = errcode.NewCommon("发送过于频繁，请稍后再试")
	// ErrVerifyTooFrequently 验证过于频繁错误
	ErrVerifyTooFrequently = errcode.NewCommon("验证过于频繁，请稍后再试")

	// ErrEmailSupport 暂不支持邮件通知服务错误
	ErrEmailSupport = errcode.NewCommon("暂不支持邮件通知服务")
	// ErrSmsSupport 暂不支持短信通知服务错误
	ErrSmsSupport = errcode.NewCommon("暂不支持短信通知服务")

	// ErrCaptchaNotExist 验证码不存在或已过期错误
	ErrCaptchaNotExist = errcode.NewCommon("验证码不存在或已过期")
	// ErrCaptchaVerify 验证码校验失败错误
	ErrCaptchaVerify = errcode.NewCommon("验证码校验失败")
)
