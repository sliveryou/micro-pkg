package notify

import (
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"

	notifytypes "github.com/sliveryou/micro-pkg/notify/types"
	"github.com/sliveryou/micro-pkg/xkv"
)

var (
	s1, _ = miniredis.Run()
	s2, _ = miniredis.Run()

	store = getStore()
)

func getStore() *xkv.Store {
	s1.FlushAll()
	s2.FlushAll()

	return xkv.NewStore([]cache.NodeConf{
		{
			RedisConf: redis.RedisConf{
				Host: s1.Addr(),
				Type: "node",
			},
			Weight: 100,
		},
		{
			RedisConf: redis.RedisConf{
				Host: s2.Addr(),
				Type: "node",
			},
			Weight: 100,
		},
	})
}

func getNotify() (*Notify, error) {
	mockClient := &notifytypes.MockClient{}

	sp := notifytypes.NewSmsClientPicker()
	sp.Add("mock-sms-1", mockClient)

	ep := notifytypes.NewEmailClientPicker()
	ep.Add("mock-email-1", mockClient)

	c := Config{
		Provider: "test",
	}

	return NewNotify(c, sp, ep, store)
}

func TestNotify_SendSmsCode(t *testing.T) {
	n, err := getNotify()
	require.NoError(t, err)

	p := &notifytypes.SendParams{}
	p.IP = "127.0.0.1"
	p.Provider = "test"
	p.Receiver = "13000000000"
	p.TemplateID = "login"
	p.Params = []notifytypes.Param{
		&notifytypes.CodeParam{Key: "code", Length: 6, Expiration: 5 * time.Minute},
		&notifytypes.CommonParam{Key: "time", Value: "5"},
	}
	p.IsMock = false

	err = n.SendSmsCode(p)
	require.NoError(t, err)
}

func TestNotify_SendEmailCode(t *testing.T) {
	n, err := getNotify()
	require.NoError(t, err)

	p := &notifytypes.SendParams{}
	p.IP = "127.0.0.1"
	p.Provider = "test"
	p.Receiver = "sliveryou@outlook.com"
	p.TemplateID = "login"
	p.Params = []notifytypes.Param{
		&notifytypes.CodeParam{Key: "code", Length: 6, Expiration: 5 * time.Minute},
		&notifytypes.CommonParam{Key: "time", Value: "5"},
	}
	p.IsMock = false

	err = n.SendEmailCode(p)
	require.NoError(t, err)
}

func TestNotify_VerifySmsCode(t *testing.T) {
	n, err := getNotify()
	require.NoError(t, err)

	p := &notifytypes.VerifyParams{}
	p.NotifyMethod = notifytypes.Sms
	p.IP = "127.0.0.1"
	p.Provider = "test"
	p.Receiver = "13000000000"
	p.TemplateID = "login"

	key := notifytypes.GenCodeKey(p.CommonParams)
	code, err := store.Get(key)
	require.NoError(t, err)

	p.Code = code
	p.Clear = false

	err = n.VerifySmsCode(p)
	require.NoError(t, err)
}

func TestNotify_VerifyEmailCode(t *testing.T) {
	n, err := getNotify()
	require.NoError(t, err)

	p := &notifytypes.VerifyParams{}
	p.NotifyMethod = notifytypes.Email
	p.IP = "127.0.0.1"
	p.Provider = "test"
	p.Receiver = "sliveryou@outlook.com"
	p.TemplateID = "login"

	key := notifytypes.GenCodeKey(p.CommonParams)
	code, err := store.Get(key)
	require.NoError(t, err)

	p.Code = code
	p.Clear = false

	err = n.VerifyEmailCode(p)
	require.NoError(t, err)
}
