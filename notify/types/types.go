package types

// Param 参数接口
type Param interface {
	GetKey() string   // 获取参数键
	GetValue() string // 获取参数值
}

// Client 客户端接口
type Client interface {
	// Platform 服务平台
	Platform() string
}

// SmsClient 短信客户端接口
type SmsClient interface {
	Client
	// SendSms 发送短信
	SendSms(receiver, templateID string, params ...Param) error
}

// EmailClient 邮件客户端接口
type EmailClient interface {
	Client
	// SendEmail 发送邮件
	SendEmail(receiver, templateID string, params ...Param) error
}

// SmsClientPicker 短信客户端选取器接口
type SmsClientPicker interface {
	// Pick 选取一个短信客户端
	Pick() (sc SmsClient, key string, isExist bool)
	// Get 获取一个短信客户端
	Get(key string) (sc SmsClient, isExist bool)
	// Add 添加一个短信客户端
	Add(key string, value SmsClient)
	// Remove 移除一个短信客户端
	Remove(keys ...string)
}

// EmailClientPicker 邮件客户端选取器接口
type EmailClientPicker interface {
	// Pick 选取一个邮件客户端
	Pick() (ec EmailClient, key string, isExist bool)
	// Get 获取一个邮件客户端
	Get(key string) (ec EmailClient, isExist bool)
	// Add 添加一个邮件客户端
	Add(key string, value EmailClient)
	// Remove 移除一个邮件客户端
	Remove(keys ...string)
}

// Params 参数列表
type Params []Param

// ToMap 将参数列表转换成map
func (ps Params) ToMap() map[string]string {
	m := make(map[string]string)

	for _, p := range ps {
		m[p.GetKey()] = p.GetValue()
	}

	return m
}

// Keys 获取参数列表所有键
func (ps Params) Keys() []string {
	keys := make([]string, 0, len(ps))

	for _, p := range ps {
		keys = append(keys, p.GetKey())
	}

	return keys
}

// Values 获取参数列表所有值
func (ps Params) Values() []string {
	values := make([]string, 0, len(ps))

	for _, p := range ps {
		values = append(values, p.GetValue())
	}

	return values
}
