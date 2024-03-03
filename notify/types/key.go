package types

import "fmt"

const (
	// KeyPrefixCode 验证码缓存 key 前缀
	KeyPrefixCode = "micro.pkg:notify:code:"
	// KeyPrefixSendLimit 发送通知限制缓存 key 前缀
	KeyPrefixSendLimit = "micro.pkg:notify:send.limit:"
	// KeyPrefixVerifyLimit 验证通知限制缓存 key 前缀
	KeyPrefixVerifyLimit = "micro.pkg:notify:verify.limit:"
	// KeyPrefixReceiverLimit 接收方限制缓存 key 前缀
	KeyPrefixReceiverLimit = "micro.pkg:notify:receiver.limit:"
	// KeyPrefixIPSourceLimit IP 来源限制缓存 key 前缀
	KeyPrefixIPSourceLimit = "micro.pkg:notify:ip.source.limit:"
	// KeyPrefixProviderLimit 提供方限制缓存 key 前缀
	KeyPrefixProviderLimit = "micro.pkg:notify:provider.limit:"
)

// GenCodeKey 生成验证码缓存 key
func GenCodeKey(p CommonParams) string {
	return fmt.Sprintf("%s%s:%s:%s:%s", KeyPrefixCode,
		p.Provider, p.NotifyMethod, p.TemplateID, p.Receiver)
}

// GenSendLimitKey 生成发送通知限制缓存 key
func GenSendLimitKey(p CommonParams) string {
	return fmt.Sprintf("%s%s:%s:%s:%s", KeyPrefixSendLimit,
		p.Provider, p.NotifyMethod, p.TemplateID, p.Receiver)
}

// GenVerifyLimitKey 生成验证通知限制缓存 key
func GenVerifyLimitKey(p CommonParams) string {
	return fmt.Sprintf("%s%s:%s:%s:%s", KeyPrefixVerifyLimit,
		p.Provider, p.NotifyMethod, p.TemplateID, p.Receiver)
}

// GenReceiverLimitKey 生成接收方限制缓存 key
func GenReceiverLimitKey(p CommonParams) string {
	return fmt.Sprintf("%s%s:%s:%s", KeyPrefixReceiverLimit,
		p.Provider, p.NotifyMethod, p.Receiver)
}

// GenIPSourceLimitKey 生成IP地址来源限制缓存 key
func GenIPSourceLimitKey(p CommonParams) string {
	return fmt.Sprintf("%s%s:%s", KeyPrefixIPSourceLimit,
		p.Provider, p.IP)
}

// GenProviderLimitKey 生成提供方限制缓存 key
func GenProviderLimitKey(p CommonParams) string {
	return fmt.Sprintf("%s%s", KeyPrefixProviderLimit,
		p.Provider)
}
