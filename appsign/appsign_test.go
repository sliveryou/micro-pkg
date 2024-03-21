package appsign

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sign "github.com/sliveryou/aliyun-api-gateway-sign"

	"github.com/sliveryou/micro-pkg/xhttp"
)

func getRawURL() string {
	rawURL := "https://test.com/api/auth"
	values := make(url.Values)
	values.Add("bankcard", "123456")
	values.Add("idcard", "123456")
	values.Add("mobile", "123456")
	values.Add("name", "测试")

	return rawURL + "?" + values.Encode()
}

func TestFromRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, getRawURL(), nil)
	err := sign.Sign(r, "123456", "123456")
	require.NoError(t, err)

	as, err := FromRequest(r)
	require.NoError(t, err)
	fmt.Println(as.StringToSign)

	signature, ok := as.CheckSign("123456")
	assert.True(t, ok)
	fmt.Println(signature)

	signature, ok = as.CheckSign("123")
	assert.False(t, ok)
	fmt.Println(signature)
}

func TestFromRequest2(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, getRawURL(), strings.NewReader(`a=1&b=2`))
	r.Header.Set(xhttp.HeaderContentType, xhttp.MIMEForm)
	err := sign.Sign(r, "123456", "123456")
	require.NoError(t, err)

	as, err := FromRequest(r)
	require.NoError(t, err)
	fmt.Println(as.StringToSign)

	signature, ok := as.CheckSign("123456")
	assert.True(t, ok)
	fmt.Println(signature)

	signature, ok = as.CheckSign("123")
	assert.False(t, ok)
	fmt.Println(signature)
}

func TestFromRequest3(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, getRawURL(), strings.NewReader(`{"a":1,"b":2}`))
	r.Header.Set(xhttp.HeaderContentType, xhttp.MIMEApplicationJSON)
	err := sign.Sign(r, "123456", "123456")
	require.NoError(t, err)

	as, err := FromRequest(r)
	require.NoError(t, err)
	fmt.Println(as.StringToSign)

	signature, ok := as.CheckSign("123456")
	assert.True(t, ok)
	fmt.Println(signature)

	signature, ok = as.CheckSign("123")
	assert.False(t, ok)
	fmt.Println(signature)
}
