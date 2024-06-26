package submail

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	smclient "github.com/sliveryou/submail-go-sdk/client"

	notifytypes "github.com/sliveryou/micro-pkg/notify/types"
	"github.com/sliveryou/micro-pkg/xhttp"
)

func getSubmail() (*Submail, error) {
	c := Config{
		Sms: &App{
			AppID:    "appID",
			AppKey:   "appKey",
			SignType: "sha1",
		},
		Email: &App{
			AppID:    "appID",
			AppKey:   "appKey",
			SignType: "sha1",
		},
	}

	httpClient := xhttp.NewHTTPClient()
	sTmpl := map[string]string{"login": "1s3mF2"}
	eTmpl := map[string]string{"login": "imvEp3"}

	return NewSubmail(c,
		notifytypes.WithHTTPClient(httpClient),
		notifytypes.WithSmsTmplMap(sTmpl),
		notifytypes.WithEmailTmplMap(eTmpl))
}

func TestMustNewSubmail(t *testing.T) {
	assert.NotPanics(t, func() {
		c := Config{
			Sms: &App{
				AppID:  "appID",
				AppKey: "appKey",
			},
			Email: &App{
				AppID:  "appID",
				AppKey: "appKey",
			},
		}

		a := MustNewSubmail(c)
		assert.NotNil(t, a)
		assert.Equal(t, smclient.SignTypeSha1, a.c.Sms.SignType)
		assert.Equal(t, smclient.SignTypeSha1, a.c.Email.SignType)
	})

	assert.NotPanics(t, func() {
		c := Config{
			Sms: &App{
				AppID:  "appID",
				AppKey: "appKey",
			},
		}

		a := MustNewSubmail(c)
		assert.NotNil(t, a)
		assert.NotNil(t, a.smsClient)
		assert.Nil(t, a.emailClient)
	})
}

func TestSubmail_SendSmsCode(t *testing.T) {
	s, err := getSubmail()
	require.NoError(t, err)
	err = s.SendSms("receiver", "login",
		&notifytypes.CodeParam{Key: "code", Value: "123456"},
		&notifytypes.CommonParam{Key: "time", Value: "5"},
	)
	t.Log(err)
}

func TestSubmail_SendEmailCode(t *testing.T) {
	s, err := getSubmail()
	require.NoError(t, err)
	err = s.SendEmail("receiver", "login",
		&notifytypes.CodeParam{Key: "code", Value: "123456"},
		&notifytypes.CommonParam{Key: "time", Value: "5"},
	)
	t.Log(err)
}
