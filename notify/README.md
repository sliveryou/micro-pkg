# 通用通知服务包 notify

通用通知服务包，包含短信、邮件验证码发送与短信、邮件验证码校验等功能，
并支持对发送间隔、验证间隔、一天内同一接收方、一天内同一 IP 和一天内总发送量的限制与监控。
   - authored by sliveryou

## 支持服务

- 赛邮云：[短信服务](https://www.mysubmail.com/sms) 和 [邮件服务](https://www.mysubmail.com/mail)
- 阿里云：[短信服务](https://help.aliyun.com/zh/sms) 和 [邮件推送](https://help.aliyun.com/product/29412.html?spm=a2c4g.29424.0.0.3c841ac0I4APvR)
- 云片：[国内短信](https://www.yunpian.com/product/domestic-sms)

## 支持功能

一个好的通知服务，我认为需要有以下几个功能：

1. 独立和通用的发送短信、邮件验证码接口，调用方只需要传递模板编号和相关参数即可；
2. 支持测试环境，在测试环境下不真正发送通知，可以发送钉钉消息或记录日志，以节约费用；
3. 支持验证码验证功能，这样调用方就不用自己缓存验证码；
4. 所有通知发送都要有记录，方便排查问题和对账；
5. 支持对一天内同一接收方、一天内同一 IP 和一天内总发送量进行限制与监控，防止通知服务被恶意盗刷。

所以通用通知服务包也是围绕这几个核心功能进行构建的。

## 设计思路

以一个调用方的角度思考，首先是对整个通知调用需要设置配额，如发送时间段内发送配额、验证时间段内验证配额、  
一天内同一接收方配额、一天内同一 IP 来源配额和一天内该提供方配额，当有一项配额超标时，返回对应错误，并不提供通知服务。  
配额相关控制逻辑是基于 redis 加载 lua 限流脚本来实现一个时间段限流器。

之后，调用方调用通知服务发送通知时，具体使用哪个第三方服务是不需要知道的，它只需要知道这样调用就能发送通知就行了，  
所以实现了 SmsClientPicker 短信客户端选取器接口和 EmailClientPicker 邮件客户端选取器接口，  
这样设计的好处是通知服务可以绑定多个发送客户端，如短信发送客户端，在准备发送短信时，随机选取一个短信发送客户端并发送，  
也可以在某个短信发送客户端经常发送失败后将其删除，使其更加高可用。比如通知服务为某个调用方准备了阿里云和赛邮云的短信发送客户端，  
调用方调用通知服务准备发送短信时，通知服务随机从阿里云和赛邮云短信发送客户端选取一个进行发送，  
当阿里云短信发送客户端出错次数较多时，可将其暂时屏蔽，只调用赛邮云客户端，保证短信发送功能的正常运行。

配置：

```go
// Config 通知服务配置
type Config struct {
	Provider      string // 提供方
	SendPeriod    int    `json:",default=60"`    // 发送时间段（与发送配额搭配，如发送时间段为 60，发送配额为 1，表示 60s 内对同一接收方只允许发送 1 次）
	SendQuota     int    `json:",default=1"`     // 发送时间段内发送配额
	VerifyPeriod  int    `json:",default=60"`    // 验证时间段（与验证配额搭配，如验证时间段为 60，验证配额为 1，表示 60s 内对同一接收方只允许验证 1 次）
	VerifyQuota   int    `json:",default=3"`     // 验证时间内段验证配额
	ReceiverQuota int    `json:",default=15"`    // 一天内同一接收方配额
	IPSourceQuota int    `json:",default=30"`    // 一天内同一IP来源配额
	ProviderQuota int    `json:",default=10000"` // 一天内该提供方配额
}
```

暴露接口：

```go
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

// Notify 通知服务
type Notify struct {
	c            Config                        // 配置
	smsClients   notifytypes.SmsClientPicker   // 短信客户端选取器
	emailClients notifytypes.EmailClientPicker // 邮件客户端选取器
	kvStore      *xkv.Store                    // 键值存取器
	periodLimit  *limit.PeriodLimit            // 通知限流器
}

// SendSmsCode 发送短信验证码
func (n *Notify) SendSmsCode(p *notifytypes.SendParams) error
// SendEmailCode 发送邮件验证码
func (n *Notify) SendEmailCode(p *notifytypes.SendParams) error
// VerifySmsCode 校验短信验证码
func (n *Notify) VerifySmsCode(p *notifytypes.VerifyParams) error
// VerifyEmailCode 校验邮箱验证码
func (n *Notify) VerifyEmailCode(p *notifytypes.VerifyParams) error
```
