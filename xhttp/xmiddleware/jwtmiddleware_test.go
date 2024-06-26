package xmiddleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/jwt"
	"github.com/sliveryou/micro-pkg/xhttp"
)

func getJWTMiddleware() *JWTMiddleware {
	j := jwt.MustNewJWT(jwt.Config{Issuer: "test-issuer", SecretKey: "ABCDEFGH", Expiration: 72 * time.Hour})
	m := MustNewJWTMiddleware(j, &userInfo{})
	return m
}

func TestJWTMiddleware_Handle_Success(t *testing.T) {
	m := getJWTMiddleware()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)

	tokenStr, err := m.j.GenToken(getToken())
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)
	req.Header.Set(xhttp.HeaderAuthorization, "Bearer "+tokenStr)

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("JWTMiddleware Handle request")

		ctx := r.Context()
		token := &userInfo{}
		err := jwt.ReadCtx(ctx, token)
		require.NoError(t, err)

		xhttp.OkJsonCtx(ctx, w, token)
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

func TestJWTMiddleware_Handle_Fail(t *testing.T) {
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
	assert.Equal(t, "{\"code\":153,\"msg\":\"Token 错误\"}", string(d))
}

func getToken() *userInfo {
	return &userInfo{
		UserID:   100000,
		UserName: "test_user",
		RoleIDs:  []int64{100000, 100001, 100002},
		Group:    "ADMIN",
		IsAdmin:  true,
		Score:    123.123,
	}
}

type userInfo struct {
	UserID   int64   `json:"user_id"`
	UserName string  `json:"user_name"`
	RoleIDs  []int64 `json:"role_ids"`
	Group    string  `json:"group"`
	IsAdmin  bool    `json:"is_admin"`
	Score    float64 `json:"score"`
}
