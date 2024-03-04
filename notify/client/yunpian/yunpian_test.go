package yunpian

import (
	"testing"

	"github.com/stretchr/testify/require"

	notifytypes "github.com/sliveryou/micro-pkg/notify/types"
	"github.com/sliveryou/micro-pkg/xhttp"
)

func getYunPian() (*YunPian, error) {
	c := Config{
		Sms: App{
			IsDisabled: false,
			APIKey:     "apiKey",
		},
	}

	httpClient := xhttp.NewHTTPClient()
	sTmpl := map[string]string{"login": "123456789"}

	return NewYunPian(c,
		notifytypes.WithHTTPClient(httpClient),
		notifytypes.WithSmsTmplMap(sTmpl),
	)
}

func TestYunPian_SendSmsCode(t *testing.T) {
	s, err := getYunPian()
	require.NoError(t, err)
	err = s.SendSms("receiver", "login",
		&notifytypes.CodeParam{Key: "code", Value: "123456"},
		&notifytypes.CommonParam{Key: "time", Value: "5"},
	)
	t.Log(err)
}
