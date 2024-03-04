package aliyun

import (
	"testing"

	"github.com/stretchr/testify/require"

	notifytypes "github.com/sliveryou/micro-pkg/notify/types"
	"github.com/sliveryou/micro-pkg/xhttp"
)

func getAliyun() (*Aliyun, error) {
	c := Config{
		Sms: App{
			RegionID:        "cn-hangzhou",
			AccessKeyID:     "accessKeyID",
			AccessKeySecret: "accessKeySecret",
			SignName:        "测试",
		},
		Email: App{
			RegionID:        "cn-hangzhou",
			AccessKeyID:     "accessKeyID",
			AccessKeySecret: "accessKeySecret",
			SignName:        "测试",
			AccountName:     "sliveryou@outlook.com",
		},
	}

	httpClient := xhttp.NewHTTPClient()
	sTmpl := map[string]string{"login": "SMS_123456789"}

	return NewAliyun(c,
		notifytypes.WithHTTPClient(httpClient),
		notifytypes.WithSmsTmplMap(sTmpl),
	)
}

func TestAliyun_SendSmsCode(t *testing.T) {
	s, err := getAliyun()
	require.NoError(t, err)
	err = s.SendSms("receiver", "login",
		&notifytypes.CodeParam{Key: "code", Value: "123456"},
		&notifytypes.CommonParam{Key: "time", Value: "5"},
	)
	t.Log(err)
}

func TestAliyun_SendEmailCode(t *testing.T) {
	s, err := getAliyun()
	require.NoError(t, err)

	eem := map[string]EmailExtra{
		"login": {
			Subject:  "登录邮件",
			TextBody: "您的登录验证码是：${code}（${time}分钟内有效），请勿泄漏给他人。如非本人操作，请忽略本条消息！",
		},
	}
	s.LoadEmailExtraMap(eem)

	err = s.SendEmail("receiver", "login",
		&notifytypes.CodeParam{Key: "code", Value: "123456"},
		&notifytypes.CommonParam{Key: "time", Value: "5"},
	)
	t.Log(err)
}
