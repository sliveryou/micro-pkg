package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSendParams_IsValid(t *testing.T) {
	var sp *SendParams
	assert.False(t, sp.IsValid())

	sp = &SendParams{}
	sp.NotifyMethod = Sms
	sp.IP = "127.0.0.1"
	sp.Provider = "test"
	sp.Receiver = "13000000000"
	sp.TemplateID = ""
	assert.False(t, sp.IsValid())

	sp.TemplateID = "login"
	sp.Params = []Param{
		&CommonParam{Key: "time", Value: "5"},
	}
	assert.True(t, sp.IsValid())

	sp.Params = []Param{
		&CodeParam{Key: "code", Length: 6, Expiration: 5 * time.Minute},
		&CommonParam{Key: "time", Value: "5"},
	}
	assert.True(t, sp.IsValid())

	ps := Params(sp.Params)
	assert.Equal(t, map[string]string{"code": "", "time": "5"}, ps.ToMap())
	assert.Equal(t, []string{"code", "time"}, ps.Keys())
	assert.Equal(t, []string{"", "5"}, ps.Values())
}

func TestVerifyParams_IsValid(t *testing.T) {
	var vp *VerifyParams
	assert.False(t, vp.IsValid())

	vp = &VerifyParams{}
	vp.NotifyMethod = Email
	vp.IP = "127.0.0.1"
	vp.Provider = "test"
	vp.Receiver = "sliveryou@outlook.com"
	vp.TemplateID = "login"
	vp.Code = ""
	assert.False(t, vp.IsValid())

	vp.Code = "code"
	assert.True(t, vp.IsValid())
}

func TestCodeParam_IsEmpty(t *testing.T) {
	cp := &CodeParam{}
	assert.True(t, cp.IsEmpty())

	cp.Key = "code"
	assert.False(t, cp.IsEmpty())
}
