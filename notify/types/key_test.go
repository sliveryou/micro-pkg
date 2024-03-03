package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenKey(t *testing.T) {
	p := CommonParams{
		NotifyMethod: Email,
		IP:           "127.0.0.1",
		Provider:     "test",
		Receiver:     "sliveryou@outlook.com",
		TemplateID:   "login",
	}

	assert.Equal(t, "micro.pkg:notify:code:test:email:login:sliveryou@outlook.com", GenCodeKey(p))
	assert.Equal(t, "micro.pkg:notify:send.limit:test:email:login:sliveryou@outlook.com", GenSendLimitKey(p))
	assert.Equal(t, "micro.pkg:notify:verify.limit:test:email:login:sliveryou@outlook.com", GenVerifyLimitKey(p))
	assert.Equal(t, "micro.pkg:notify:receiver.limit:test:email:sliveryou@outlook.com", GenReceiverLimitKey(p))
	assert.Equal(t, "micro.pkg:notify:ip.source.limit:test:127.0.0.1", GenIPSourceLimitKey(p))
	assert.Equal(t, "micro.pkg:notify:provider.limit:test", GenProviderLimitKey(p))
}
