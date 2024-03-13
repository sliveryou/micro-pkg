package corpaccount

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCorpAccount_Authenticate(t *testing.T) {
	c := Config{IsMock: false, AppKey: "", AppKeySecret: "appKeySecret"}
	_, err := NewCorpAccount(c)
	require.EqualError(t, err, "corpaccount: illegal corpaccount config")

	c = Config{IsMock: false, AppKey: "appKey", AppKeySecret: ""}
	_, err = NewCorpAccount(c)
	require.EqualError(t, err, "corpaccount: illegal corpaccount config")

	c = Config{IsMock: false, AppKey: "appKey", AppKeySecret: "appKeySecret"}
	ca, err := NewCorpAccount(c)
	require.NoError(t, err)

	req := &AuthenticateRequest{
		AcctName: "杭州测试有限公司",
		BankName: "建设银行",
		CardNo:   "33050161963500000428",
	}
	resp, err := ca.Authenticate(context.Background(), req)
	t.Log(resp, err)
}
