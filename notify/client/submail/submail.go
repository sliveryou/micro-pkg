package submail

import (
	"github.com/pkg/errors"

	smclient "github.com/sliveryou/submail-go-sdk/client"
	"github.com/sliveryou/submail-go-sdk/mail"
	"github.com/sliveryou/submail-go-sdk/sms"

	notifytypes "github.com/sliveryou/micro-pkg/notify/types"
)

const (
	// PlatformSubmail 赛邮云通知平台
	PlatformSubmail = "submail"
)

// App 应用相关配置
type App struct {
	IsDisabled bool   `json:",optional"`                               // 是否禁用
	AppID      string `json:",optional"`                               // 应用ID
	AppKey     string `json:",optional"`                               // 应用Key
	SignType   string `json:",default=sha1,options=[normal,sha1,md5]"` // 签名类型（枚举 normal、sha1 和 md5）
}

// Config 赛邮云通知服务相关配置
type Config struct {
	Sms   App // 短信应用相关配置
	Email App // 邮件应用相关配置
}

// Submail 赛邮云通知服务结构详情
type Submail struct {
	c           Config                  // 相关配置
	baseClient  *notifytypes.BaseClient // 基础客户端
	smsClient   *sms.Client             // 短信客户端
	emailClient *mail.Client            // 邮件客户端
}

// NewSubmail 新建赛邮云通知服务对象
func NewSubmail(c Config, opts ...notifytypes.Option) (*Submail, error) {
	if err := c.check(); err != nil {
		return nil, errors.WithMessage(err, "submail: check config err")
	}

	s := &Submail{
		c:          c,
		baseClient: notifytypes.NewBaseClient(c.Sms.IsDisabled, c.Email.IsDisabled, opts...),
	}

	if !c.Sms.IsDisabled {
		s.smsClient = sms.New(c.Sms.AppID, c.Sms.AppKey, c.Sms.SignType,
			smclient.WithHTTPClient(s.baseClient.HTTPClient))
	}

	if !c.Email.IsDisabled {
		s.emailClient = mail.New(c.Email.AppID, c.Email.AppKey, c.Email.SignType,
			smclient.WithHTTPClient(s.baseClient.HTTPClient))
	}

	return s, nil
}

// MustNewSubmail 新建赛邮云通知服务对象
func MustNewSubmail(c Config, opts ...notifytypes.Option) *Submail {
	s, err := NewSubmail(c, opts...)
	if err != nil {
		panic(err)
	}

	return s
}

// Platform 服务平台
func (s *Submail) Platform() string {
	return PlatformSubmail
}

// SendSms 发送短信
func (s *Submail) SendSms(receiver, templateID string, params ...notifytypes.Param) error {
	parsed, err := s.baseClient.ParseSmsTmpl(templateID)
	if err != nil {
		return err
	}

	xsp := &sms.XSendParam{
		To:      receiver,
		Project: parsed,
		Vars:    notifytypes.Params(params).ToMap(),
	}

	return errors.WithMessage(s.smsClient.XSend(xsp), "sms client xsend err")
}

// SendEmail 发送邮件
func (s *Submail) SendEmail(receiver, templateID string, params ...notifytypes.Param) error {
	parsed, err := s.baseClient.ParseEmailTmpl(templateID)
	if err != nil {
		return err
	}

	xsp := &mail.XSendParam{
		To:           []*mail.ToParam{{Address: receiver}},
		Project:      parsed,
		Vars:         notifytypes.Params(params).ToMap(),
		Asynchronous: false,
	}

	return errors.WithMessage(s.emailClient.XSend(xsp), "email client xsend err")
}

// isValid 判断应用相关配置是否合法
func (a *App) isValid() bool {
	if !a.IsDisabled {
		if a.AppID == "" || a.AppKey == "" {
			return false
		}
		if a.SignType == "" {
			a.SignType = smclient.SignTypeSha1
		}
	}

	return true
}

// check 检查配置
func (c *Config) check() error {
	if !c.Sms.isValid() {
		return errors.New("illegal submail sms config")
	}

	if !c.Email.isValid() {
		return errors.New("illegal submail email config")
	}

	return nil
}
