package xhttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"

	"github.com/sliveryou/micro-pkg/disabler"
	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/jwt"
)

func getJWTMiddleware() *JWTMiddleware {
	j := jwt.MustNewJWT(jwt.Config{Issuer: "test-issuer", SecretKey: "ABCDEFGH", Expiration: 72 * time.Hour})
	errTokenVerify := errcode.New(401, "token校验失败", http.StatusUnauthorized)
	m := MustNewJWTMiddleware(j, &_UserInfo{}, errTokenVerify)
	return m
}

func TestNewJWTMiddleware_Success(t *testing.T) {
	m := getJWTMiddleware()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)

	tokenStr, err := m.j.GenToken(getToken())
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)
	req.Header.Set(HeaderAuthorization, "Bearer "+tokenStr)

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("JWTMiddleware Handle request")

		ctx := r.Context()
		token := &_UserInfo{}
		err := jwt.ReadCtx(ctx, token)
		require.NoError(t, err)

		OkJsonCtx(ctx, w, token)
	})
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	result := resp.Result()
	defer result.Body.Close()
	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, "{\"code\":0,\"msg\":\"ok\",\"data\":{\"user_id\":100000,\"user_name\":\"test_user\",\"role_ids\":[100000,100001,100002],\"group\":\"ADMIN\",\"is_admin\":true,\"score\":123.123}}", string(d))
}

func TestNewJWTMiddleware_Fail(t *testing.T) {
	m := getJWTMiddleware()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {})
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	result := resp.Result()
	defer result.Body.Close()
	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, "{\"code\":401,\"msg\":\"token校验失败\"}", string(d))
}

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
		OkJsonCtx(r.Context(), w, nil)
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

func TestNewCorsMiddleware(t *testing.T) {
	m := NewCorsMiddleware()
	req := httptest.NewRequest(http.MethodOptions, "http://localhost", http.NoBody)
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("CorsMiddleware Handle request")
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNoContent, resp.Code)
	t.Log(resp.Header())
}

func TestNewRecoverMiddleware(t *testing.T) {
	m := NewRecoverMiddleware()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("RecoverMiddleware Handle request")
		panic(errcode.ErrUnexpected)
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	t.Log(resp.Header())
}

func TestNewRLogMiddleware(t *testing.T) {
	m := NewRLogMiddleware()
	req := httptest.NewRequest(http.MethodPost, "http://localhost", strings.NewReader(`{"id":1}`))
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("RLogMiddleware Handle request")

		var buf bytes.Buffer
		io.Copy(&buf, req.Body)
		t.Log(buf.String())

		w.Header().Set(HeaderContentType, ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":200,"msg":"ok"}`))
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestNewIgnoreRLogMiddleware(t *testing.T) {
	m := NewIgnoreRLogMiddleware()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("IgnoreRLogMiddleware Handle request")
		md, ok := metadata.FromOutgoingContext(r.Context())
		assert.True(t, ok)
		t.Log(md)
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func getToken() *_UserInfo {
	return &_UserInfo{
		UserID:   100000,
		UserName: "test_user",
		RoleIDs:  []int64{100000, 100001, 100002},
		Group:    "ADMIN",
		IsAdmin:  true,
		Score:    123.123,
	}
}

type _UserInfo struct {
	UserID   int64   `json:"user_id"`
	UserName string  `json:"user_name"`
	RoleIDs  []int64 `json:"role_ids"`
	Group    string  `json:"group"`
	IsAdmin  bool    `json:"is_admin"`
	Score    float64 `json:"score"`
}
