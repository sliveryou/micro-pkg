package xmiddleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/disabler"
	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/xhttp"
)

func getFuncDisable() *FuncDisableMiddleware {
	fd := disabler.MustNewFuncDisabler(disabler.Config{
		DisabledAPIs: []string{
			"GET:/api/file",
			"/api/user",
			"/api/auth/{id}",
		},
	})
	errNotAllowed := errcode.NewCommon("暂不支持该 API")
	m := MustNewFuncDisableMiddleware(fd, "", errNotAllowed)
	return m
}

func TestNewFuncDisable_Success(t *testing.T) {
	m := getFuncDisable()
	req := httptest.NewRequest(http.MethodPost, "https://test.com/api/file", http.NoBody)
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		xhttp.OkJsonCtx(r.Context(), w, nil)
	})
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	result := resp.Result()
	defer result.Body.Close()
	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, "{\"code\":0,\"msg\":\"ok\"}", string(d))
}

func TestNewFuncDisable_Fail(t *testing.T) {
	m := getFuncDisable()
	req := httptest.NewRequest(http.MethodGet, "https://test.com/api/auth/1", http.NoBody)
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {})
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	result := resp.Result()
	defer result.Body.Close()
	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, "{\"code\":97,\"msg\":\"暂不支持该 API\"}", string(d))
}
