package xmiddleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"

	sign "github.com/sliveryou/aliyun-api-gateway-sign"
	"github.com/sliveryou/go-tool/v2/convert"

	"github.com/sliveryou/micro-pkg/appsign"
	"github.com/sliveryou/micro-pkg/xhttp"
	"github.com/sliveryou/micro-pkg/xkv"
)

var (
	s1, _ = miniredis.Run()
	s2, _ = miniredis.Run()
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

func getSignMiddleware() *SignMiddleware {
	return MustNewSignMiddleware(getStore(), func(ctx context.Context, appKey string) (appKeySecret string, err error) {
		return "appKeySecret", nil
	})
}

func TestSignMiddleware_Handle_Success(t *testing.T) {
	m := getSignMiddleware()
	req := httptest.NewRequest(http.MethodGet, getRawURL(), nil)
	err := sign.Sign(req, "appKey", "appKeySecret")
	require.NoError(t, err)

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("SignMiddleware Handle request")

		ctx := r.Context()
		appKey := appsign.AppKeyFromCtx(ctx)
		assert.NotEmpty(t, appKey)

		xhttp.OkJsonCtx(ctx, w, appKey)
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	result := resp.Result()
	defer result.Body.Close()
	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"code":0,"msg":"ok","data":"appKey"}`, string(d))
}

func TestSignMiddleware_Handle_Fail(t *testing.T) {
	// 测试错误密钥
	m := getSignMiddleware()
	req := httptest.NewRequest(http.MethodPost, getRawURL(), strings.NewReader(`{"a":1,"b":2}`))
	err := sign.Sign(req, "appKey", "appKeySecretErr")
	require.NoError(t, err)

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	result := resp.Result()
	defer result.Body.Close()
	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Contains(t, string(d), "签名错误，服务端计算的待签名字符串为")
}

func TestSignMiddleware_Handle_Fail2(t *testing.T) {
	// 测试签名过期
	m := getSignMiddleware()
	req := httptest.NewRequest(http.MethodPost, getRawURL(), strings.NewReader(`{"a":1,"b":2}`))
	err := sign.Sign(req, "appKey", "appKeySecret")
	require.NoError(t, err)
	errTimestamp := convert.ToString(time.Now().Add(-10 * time.Minute).UnixMilli())
	req.Header.Set(appsign.HeaderCATimestamp, errTimestamp)

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	result := resp.Result()
	defer result.Body.Close()
	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"code":151,"msg":"签名已过期"}`, string(d))
}

func TestSignMiddleware_Handle_Fail3(t *testing.T) {
	// 测试请求重放
	m := getSignMiddleware()
	req := httptest.NewRequest(http.MethodPost, getRawURL(), strings.NewReader(`{"a":1,"b":2}`))
	err := sign.Sign(req, "appKey", "appKeySecret")
	require.NoError(t, err)
	req2, err := xhttp.CopyRequest(req)
	require.NoError(t, err)

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("SignMiddleware Handle request")

		ctx := r.Context()
		appKey := appsign.AppKeyFromCtx(ctx)
		assert.NotEmpty(t, appKey)

		xhttp.OkJsonCtx(ctx, w, appKey)
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	result := resp.Result()
	defer result.Body.Close()
	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"code":0,"msg":"ok","data":"appKey"}`, string(d))

	resp2 := httptest.NewRecorder()
	handler.ServeHTTP(resp2, req2)
	assert.Equal(t, http.StatusUnauthorized, resp2.Code)

	result2 := resp2.Result()
	defer result2.Body.Close()
	d2, err := io.ReadAll(result2.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"code":152,"msg":"随机数已过期"}`, string(d2))
}

func getRawURL() string {
	rawURL := "https://test.com/api/auth"
	values := make(url.Values)
	values.Add("bankcard", "123456")
	values.Add("idcard", "123456")
	values.Add("mobile", "123456")
	values.Add("name", "测试")

	return rawURL + "?" + values.Encode()
}
