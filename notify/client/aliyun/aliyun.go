package aliyun

import (
	"encoding/json"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/errcode"
	notifytypes "github.com/sliveryou/micro-pkg/notify/types"
)

const (
	// PlatformAliyun 阿里云通知平台
	PlatformAliyun = "aliyun"

	// codeOK 调用成功状态码
	codeOK = "OK"
	// defaultRegionID 默认地域ID
	defaultRegionID = "cn-hangzhou"
)

// ErrEmailTmplNotExist 邮件模板不存在错误
var ErrEmailTmplNotExist = errcode.NewCommon("邮件模板信息不存在")

// App 应用相关配置
type App struct {
	IsDisabled      bool   `json:",optional"`            // 是否禁用
	RegionID        string `json:",default=cn-hangzhou"` // 地域ID
	AccessKeyID     string `json:",optional"`            // 访问鉴权ID
	AccessKeySecret string `json:",optional"`            // 访问鉴权私钥
	SignName        string `json:",optional"`            // 短信签名名称
	AccountName     string `json:",optional"`            // 发信地址（邮件应用使用）
}

// Config 阿里云通知服务相关配置
type Config struct {
	Sms   App // 短信应用相关配置
	Email App // 邮件应用相关配置
}

// Aliyun 阿里云通知服务结构详情
type Aliyun struct {
	c             Config                  // 相关配置
	baseClient    *notifytypes.BaseClient // 基础客户端
	smsClient     *dysmsapi.Client        // 短信客户端
	emailClient   *sdk.Client             // 邮件客户端
	emailExtraMap map[string]EmailExtra   // 邮件额外信息映射
}

// NewAliyun 新建阿里云通知服务对象
func NewAliyun(c Config, opts ...notifytypes.Option) (*Aliyun, error) {
	if err := c.check(); err != nil {
		return nil, errors.WithMessage(err, "aliyun: check config err")
	}

	a := &Aliyun{
		c:             c,
		baseClient:    notifytypes.NewBaseClient(c.Sms.IsDisabled, c.Email.IsDisabled, opts...),
		emailExtraMap: make(map[string]EmailExtra),
	}

	if !c.Sms.IsDisabled {
		config := sdk.NewConfig()
		config.Transport = a.baseClient.HTTPClient.Transport
		config.Timeout = a.baseClient.HTTPClient.Timeout
		credential := &credentials.AccessKeyCredential{AccessKeyId: c.Sms.AccessKeyID, AccessKeySecret: c.Sms.AccessKeySecret}
		smsClient, err := dysmsapi.NewClientWithOptions(c.Sms.RegionID, config, credential)
		if err != nil {
			return nil, errors.WithMessage(err, "aliyun: new sms client err")
		}

		a.smsClient = smsClient
	}

	if !c.Email.IsDisabled {
		config := sdk.NewConfig()
		config.Transport = a.baseClient.HTTPClient.Transport
		config.Timeout = a.baseClient.HTTPClient.Timeout
		credential := &credentials.AccessKeyCredential{AccessKeyId: c.Email.AccessKeyID, AccessKeySecret: c.Email.AccessKeySecret}
		emailClient, err := sdk.NewClientWithOptions(c.Email.RegionID, config, credential)
		if err != nil {
			return nil, errors.WithMessage(err, "aliyun: new email client err")
		}

		a.emailClient = emailClient
	}

	return a, nil
}

// MustNewAliyun 新建阿里云通知服务对象
func MustNewAliyun(c Config, opts ...notifytypes.Option) *Aliyun {
	a, err := NewAliyun(c, opts...)
	if err != nil {
		panic(err)
	}

	return a
}

// Platform 服务平台
func (a *Aliyun) Platform() string {
	return PlatformAliyun
}

// SendSms 发送短信
func (a *Aliyun) SendSms(receiver, templateID string, params ...notifytypes.Param) error {
	parsed, err := a.baseClient.ParseSmsTmpl(templateID)
	if err != nil {
		return err
	}

	var templateParam string
	if len(params) > 0 {
		m := notifytypes.Params(params).ToMap()
		b, err := json.Marshal(&m)
		if err != nil {
			return errors.WithMessage(err, "json marshal params err")
		}

		templateParam = string(b)
	}

	// https://api.aliyun.com/document/Dysmsapi/2017-05-25/SendSms
	req := dysmsapi.CreateSendSmsRequest()
	req.Scheme = "https"
	req.PhoneNumbers = receiver
	req.SignName = a.c.Sms.SignName
	req.TemplateCode = parsed
	req.TemplateParam = templateParam

	resp, err := a.smsClient.SendSms(req)
	if err != nil {
		return errors.WithMessage(err, "sms client send sms err")
	}

	if resp.Code != codeOK {
		return errors.New(resp.Code + ": " + resp.Message)
	}

	return nil
}

// SendEmail 发送邮件
func (a *Aliyun) SendEmail(receiver, templateID string, params ...notifytypes.Param) error {
	if a.c.Email.IsDisabled {
		return notifytypes.ErrEmailSupport
	}
	ee, ok := a.emailExtraMap[templateID]
	if !ok {
		return ErrEmailTmplNotExist
	}

	textBody := ee.TextBody
	for _, param := range params {
		textBody = strings.ReplaceAll(textBody, "${"+param.GetKey()+"}", param.GetValue())
	}

	// https://next.api.aliyun.com/document/Dm/2015-11-23/SingleSendMail
	req := requests.NewCommonRequest()
	req.Method = "POST"
	req.Scheme = "https"
	req.Domain = "dm.aliyuncs.com"
	req.Version = "2015-11-23"
	req.ApiName = "SingleSendMail"
	req.QueryParams["AccountName"] = a.c.Email.AccountName
	req.QueryParams["AddressType"] = "1"
	req.QueryParams["ReplyToAddress"] = "true"
	req.QueryParams["ToAddress"] = receiver
	req.QueryParams["Subject"] = ee.Subject
	req.QueryParams["TextBody"] = textBody
	req.QueryParams["FromAlias"] = a.c.Email.SignName

	resp, err := a.emailClient.ProcessCommonRequest(req)
	if err != nil {
		return errors.WithMessage(err, "email client send email err")
	}

	var cp commonResponse
	err = json.Unmarshal(resp.GetHttpContentBytes(), &cp)
	if err != nil {
		return errors.WithMessage(err, "json unmarshal err")
	}

	if !resp.IsSuccess() {
		return errors.New(cp.Code + ": " + cp.Message)
	}

	return nil
}

// LoadEmailExtraMap 加载邮件额外信息映射
func (a *Aliyun) LoadEmailExtraMap(eem map[string]EmailExtra) {
	if eem != nil {
		a.emailExtraMap = eem
	}
}

// isValid 判断应用相关配置是否合法
func (a *App) isValid(isEmailApp ...bool) bool {
	if !a.IsDisabled {
		if a.AccessKeyID == "" || a.AccessKeySecret == "" || a.SignName == "" {
			return false
		}
		// 如果为邮件应用且发信地址为空
		if len(isEmailApp) > 0 && isEmailApp[0] && a.AccountName == "" {
			return false
		}
		if a.RegionID == "" {
			a.RegionID = defaultRegionID
		}
	}

	return true
}

// check 检查配置
func (c *Config) check() error {
	if !c.Sms.isValid() {
		return errors.New("illegal aliyun sms config")
	}

	if !c.Email.isValid(true) {
		return errors.New("illegal aliyun email config")
	}

	return nil
}

// EmailExtra 邮件额外信息
type EmailExtra struct {
	Subject  string `json:"subject"`   // 邮件标题
	TextBody string `json:"text_body"` // 邮件text正文
}

// commonResponse 通用响应
type commonResponse struct {
	EnvID     string `json:"EnvId"`
	RequestID string `json:"RequestId"`
	HostID    string `json:"HostId"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	Recommend string `json:"Recommend"`
}
