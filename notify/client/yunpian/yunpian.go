package yunpian

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/yunpian/yunpian-go-sdk/sdk"

	notifytypes "github.com/sliveryou/micro-pkg/notify/types"
)

const (
	// PlatformYunPian 云片通知平台
	PlatformYunPian = "yunpian"
)

// App 应用相关配置
type App struct {
	IsDisabled bool   `json:",optional"` // 是否禁用
	APIKey     string `json:",optional"` // 接口Key
}

// Config 云片通知服务相关配置
type Config struct {
	Sms App // 短信应用相关配置
}

// YunPian 云片知服务结构详情
type YunPian struct {
	c          Config                  // 相关配置
	baseClient *notifytypes.BaseClient // 基础客户端
	smsClient  sdk.YunpianClient       // 短信客户端
}

// NewYunPian 新建云片通知服务对象
func NewYunPian(c Config, opts ...notifytypes.Option) (*YunPian, error) {
	y := &YunPian{c: c, baseClient: notifytypes.NewBaseClient(
		c.Sms.IsDisabled, true, opts...)}

	if !c.Sms.IsDisabled {
		if c.Sms.APIKey == "" {
			return nil, errors.New("yunpian: illegal yunpian sms config")
		}

		smsClient := sdk.New(c.Sms.APIKey)
		smsClient.WithHttp(y.baseClient.HTTPClient)

		y.smsClient = smsClient
	}

	return y, nil
}

// MustNewYunPian 新建云片通知服务对象
func MustNewYunPian(c Config, opts ...notifytypes.Option) *YunPian {
	a, err := NewYunPian(c, opts...)
	if err != nil {
		panic(err)
	}

	return a
}

// Platform 服务平台
func (y *YunPian) Platform() string {
	return PlatformYunPian
}

// SendSms 发送短信
func (y *YunPian) SendSms(receiver, templateID string, params ...notifytypes.Param) error {
	parsed, err := y.baseClient.ParseSmsTmpl(templateID)
	if err != nil {
		return err
	}

	var tplValue string
	if len(params) > 0 {
		var buf strings.Builder
		for _, param := range params {
			key := fmt.Sprintf("#%s#", param.GetKey())
			buf.WriteString(url.QueryEscape(key))
			buf.WriteString("=")
			buf.WriteString(url.QueryEscape(param.GetValue()))
			buf.WriteString("&")
		}
		tplValue = strings.TrimSuffix(buf.String(), "&")
	}

	// https://www.yunpian.com/official/document/sms/zh_CN/domestic_tpl_single_send
	p := sdk.NewParam(3)
	p[sdk.MOBILE] = receiver
	p[sdk.TPL_ID] = parsed
	p[sdk.TPL_VALUE] = tplValue

	resp := y.smsClient.Sms().TplSingleSend(p)

	switch resp.Code {
	case sdk.SUCC:
		return nil
	case sdk.UNKOWN:
		return errors.New(resp.Msg)
	default:
		return errors.New(resp.Msg + ": " + resp.Detail)
	}
}
