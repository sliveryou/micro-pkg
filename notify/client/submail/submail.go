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
	IsDisabled bool   `json:",optional"` // 是否禁用
	AppID      string // 应用id
	AppKey     string // 应用key
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
	s := &Submail{c: c, baseClient: notifytypes.NewBaseClient(
		c.Sms.IsDisabled, c.Email.IsDisabled, opts...)}

	if !c.Sms.IsDisabled {
		if c.Sms.AppID == "" || c.Sms.AppKey == "" {
			return nil, errors.New("submail: illegal submail sms config")
		}

		s.smsClient = sms.New(c.Sms.AppID, c.Sms.AppKey, c.Sms.SignType,
			smclient.WithHTTPClient(s.baseClient.HTTPClient))
	}

	if !c.Email.IsDisabled {
		if c.Email.AppID == "" || c.Email.AppKey == "" {
			return nil, errors.New("submail: illegal submail email config")
		}

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

	return s.smsClient.XSend(xsp)
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

	return s.emailClient.XSend(xsp)
}
