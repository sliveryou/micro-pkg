package notify

import (
	"time"

	"dario.cat/mergo"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/threading"

	"github.com/sliveryou/go-tool/v2/randx"

	"github.com/sliveryou/micro-pkg/limit"
	notifytypes "github.com/sliveryou/micro-pkg/notify/types"
	"github.com/sliveryou/micro-pkg/xkv"
)

// MockCode 模拟验证码
var MockCode = "123456"

// Config 通知服务相关配置
type Config struct {
	IsDisabled    bool   `json:",optional"` // 是否禁用
	Provider      string // 提供方
	SendPeriod    int    `json:",default=60"`    // 发送时间段（与发送配额搭配，如发送时间段为 60，发送配额为 1，表示 60s 内对同一接收方只允许发送 1 次）
	SendQuota     int    `json:",default=1"`     // 发送时间段内发送配额
	VerifyPeriod  int    `json:",default=60"`    // 验证时间段（与验证配额搭配，如验证时间段为 60，验证配额为 1，表示 60s 内对同一接收方只允许验证 1 次）
	VerifyQuota   int    `json:",default=3"`     // 验证时间内段验证配额
	ReceiverQuota int    `json:",default=15"`    // 一天内同一接收方配额
	IPSourceQuota int    `json:",default=30"`    // 一天内同一IP来源配额
	ProviderQuota int    `json:",default=10000"` // 一天内该提供方配额
}

// Notify 通知服务结构详情
type Notify struct {
	c            Config                        // 相关配置
	smsClients   notifytypes.SmsClientPicker   // 短信客户端选取器
	emailClients notifytypes.EmailClientPicker // 邮件客户端选取器
	kvStore      *xkv.Store                    // 键值存取器
	periodLimit  *limit.PeriodLimit            // 通知限流器
}

// NewNotify 新建通知服务对象
func NewNotify(c Config, smsClients notifytypes.SmsClientPicker, emailClients notifytypes.EmailClientPicker, kvStore *xkv.Store) (*Notify, error) {
	if smsClients == nil || emailClients == nil || kvStore == nil || c.Provider == "" {
		return nil, errors.New("notify: illegal notify config")
	}
	if err := c.fillDefault(); err != nil {
		return nil, errors.WithMessage(err, "notify: fill default config err")
	}

	// 默认限流时间段为 24*3600 秒 = 24 小时，默认限流时间段内配额为 15 次
	periodLimit, err := limit.NewPeriodLimit(
		24*3600, 15, "", kvStore)
	if err != nil {
		return nil, errors.WithMessage(err, "notify: new period limit err")
	}

	return &Notify{
		c:            c,
		smsClients:   smsClients,
		emailClients: emailClients,
		kvStore:      kvStore,
		periodLimit:  periodLimit,
	}, nil
}

// MustNewNotify 新建通知服务对象
func MustNewNotify(c Config, smsClients notifytypes.SmsClientPicker, emailClients notifytypes.EmailClientPicker, kvStore *xkv.Store) *Notify {
	n, err := NewNotify(c, smsClients, emailClients, kvStore)
	if err != nil {
		panic(err)
	}

	return n
}

// SendSmsCode 发送短信验证码
func (n *Notify) SendSmsCode(p *notifytypes.SendParams) error {
	p.NotifyMethod = notifytypes.Sms
	return n.handleSend(p)
}

// SendEmailCode 发送邮件验证码
func (n *Notify) SendEmailCode(p *notifytypes.SendParams) error {
	p.NotifyMethod = notifytypes.Email
	return n.handleSend(p)
}

// VerifySmsCode 校验短信验证码
func (n *Notify) VerifySmsCode(p *notifytypes.VerifyParams) error {
	p.NotifyMethod = notifytypes.Sms
	return n.handleVerify(p)
}

// VerifyEmailCode 校验邮箱验证码
func (n *Notify) VerifyEmailCode(p *notifytypes.VerifyParams) error {
	p.NotifyMethod = notifytypes.Email
	return n.handleVerify(p)
}

// handleSendParams 处理发送通知参数
func (n *Notify) handleSendParams(p *notifytypes.SendParams) (*notifytypes.CodeParam, error) {
	if !p.IsValid() || p.Provider != n.c.Provider {
		return nil, notifytypes.ErrInvalidParams
	}

	var cp *notifytypes.CodeParam
	for _, param := range p.Params {
		if codeParam, ok := param.(*notifytypes.CodeParam); ok {
			cp = codeParam
		}
	}

	if cp == nil {
		return &notifytypes.CodeParam{}, nil
	}
	if cp.Key == "" {
		return nil, notifytypes.ErrInvalidParams
	}
	if cp.Length <= 0 {
		cp.Length = 6
	}
	if cp.Expiration <= 0 {
		cp.Expiration = 5 * time.Minute
	}
	if cp.Value == "" {
		if p.IsMock {
			cp.Value = MockCode
		} else {
			cp.Value = randx.NewNumber(cp.Length)
		}
	}

	return cp, nil
}

// checkSend 检查给定参数条件是否允许发送
func (n *Notify) checkSend(p notifytypes.CommonParams) error {
	// 生成提供方限制缓存 key
	providerKey := notifytypes.GenProviderLimitKey(p)
	ok, err := n.periodLimit.Allow(providerKey, limit.WithQuota(n.c.ProviderQuota))
	if err != nil {
		return errors.Wrap(err, "provider limit allow err")
	}
	if !ok {
		return notifytypes.ErrProviderOverQuota
	}

	// 生成IP地址来源限制缓存 key
	ipSourceKey := notifytypes.GenIPSourceLimitKey(p)
	ok, err = n.periodLimit.Allow(ipSourceKey, limit.WithQuota(n.c.IPSourceQuota))
	if err != nil {
		return errors.Wrap(err, "ip source limit allow err")
	}
	if !ok {
		return notifytypes.ErrIPSourceOverQuota
	}

	// 生成接收方限制缓存 key
	receiverKey := notifytypes.GenReceiverLimitKey(p)
	ok, err = n.periodLimit.Allow(receiverKey, limit.WithQuota(n.c.ReceiverQuota))
	if err != nil {
		return errors.Wrap(err, "receiver limit allow err")
	}
	if !ok {
		return notifytypes.ErrReceiverOverQuota
	}

	// 生成发送通知限制缓存 key
	sendKey := notifytypes.GenSendLimitKey(p)
	ok, err = n.periodLimit.Allow(sendKey,
		limit.WithPeriod(n.c.SendPeriod), limit.WithQuota(n.c.SendQuota))
	if err != nil {
		return errors.Wrap(err, "send limit allow err")
	}
	if !ok {
		return notifytypes.ErrSendTooFrequently
	}

	return nil
}

// handleSend 处理发送通知
func (n *Notify) handleSend(p *notifytypes.SendParams) error {
	if n.c.IsDisabled {
		return notifytypes.ErrProviderSupport
	}

	// 处理发送通知参数
	cp, err := n.handleSendParams(p)
	if err != nil {
		return err
	}

	// 检查给定参数条件是否允许发送
	err = n.checkSend(p.CommonParams)
	if err != nil {
		return err
	}

	if !cp.IsEmpty() {
		key := notifytypes.GenCodeKey(p.CommonParams)
		err = n.kvStore.SetString(key, cp.Value, int(cp.Expiration.Seconds()))
		if err != nil {
			return errors.Wrapf(err, "kv store set string by "+
				"key = %v, value = %v err", key, cp.Value)
		}
	}

	if !p.IsMock {
		switch p.NotifyMethod {
		case notifytypes.Email:
			ec, key, isExist := n.emailClients.Pick()
			if !isExist {
				return notifytypes.ErrEmailSupport
			}
			return errors.Wrapf(ec.SendEmail(p.Receiver, p.TemplateID, p.Params...),
				"send email by key = %v, params = %+v err", key, p)
		default:
			sc, key, isExist := n.smsClients.Pick()
			if !isExist {
				return notifytypes.ErrSmsSupport
			}
			return errors.Wrapf(sc.SendSms(p.Receiver, p.TemplateID, p.Params...),
				"send sms by key = %v, params = %+v err", key, p)
		}
	}

	return nil
}

// handleVerifyParams 处理校验通知参数
func (n *Notify) handleVerifyParams(p *notifytypes.VerifyParams) error {
	if !p.IsValid() || p.Provider != n.c.Provider {
		return notifytypes.ErrInvalidParams
	}

	return nil
}

// checkVerify 检查给定参数条件是否允许校验
func (n *Notify) checkVerify(p notifytypes.CommonParams) error {
	// 生成验证通知限制缓存 key
	verifyKey := notifytypes.GenVerifyLimitKey(p)
	ok, err := n.periodLimit.Allow(verifyKey,
		limit.WithPeriod(n.c.VerifyPeriod), limit.WithQuota(n.c.VerifyQuota))
	if err != nil {
		return errors.Wrap(err, "verify limit allow err")
	}
	if !ok {
		return notifytypes.ErrVerifyTooFrequently
	}

	return nil
}

// handleVerify 处理校验通知
func (n *Notify) handleVerify(p *notifytypes.VerifyParams) error {
	if n.c.IsDisabled {
		return notifytypes.ErrProviderSupport
	}

	// 处理校验通知参数
	err := n.handleVerifyParams(p)
	if err != nil {
		return err
	}

	// 检查给定参数条件是否允许校验
	err = n.checkVerify(p.CommonParams)
	if err != nil {
		return err
	}

	// 生成验证码缓存 key
	key := notifytypes.GenCodeKey(p.CommonParams)
	cacheCode, err := n.kvStore.Get(key)
	if err != nil {
		return errors.Wrapf(err, "kv store get by key = %v err", key)
	}

	if cacheCode == "" {
		return notifytypes.ErrCaptchaNotExist
	}

	if p.Code != cacheCode {
		return notifytypes.ErrCaptchaVerify
	}

	if p.Clear {
		threading.GoSafe(func() { n.kvStore.Del(key) })
	}

	return nil
}

// fillDefault 填充默认值
func (c *Config) fillDefault() error {
	fill := &Config{}
	if err := conf.FillDefault(fill); err != nil {
		return err
	}

	return mergo.Merge(c, fill)
}
