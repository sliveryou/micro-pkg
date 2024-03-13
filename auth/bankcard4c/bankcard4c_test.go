package bankcard4c

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBankCard4C_Authenticate(t *testing.T) {
	c := Config{IsMock: false, AppKey: "", AppKeySecret: "appKeySecret"}
	_, err := NewBankCard4C(c)
	require.EqualError(t, err, "bankcard4c: illegal bankcard4c config")

	c = Config{IsMock: false, AppKey: "appKey", AppKeySecret: ""}
	_, err = NewBankCard4C(c)
	require.EqualError(t, err, "bankcard4c: illegal bankcard4c config")

	c = Config{IsMock: false, AppKey: "appKey", AppKeySecret: "appKeySecret"}
	b, err := NewBankCard4C(c)
	require.NoError(t, err)

	req := &AuthenticateRequest{
		BankCard: "1234567891234567890",
		IDCard:   "330333199001053317",
		Mobile:   "12345678910",
		Name:     "陈测试",
	}
	resp, err := b.Authenticate(context.Background(), req)
	t.Log(resp, err)
}
