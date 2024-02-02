package jwt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/sliveryou/go-tool/v2/timex"
)

func TestJWT_GenToken(t *testing.T) {
	c := Config{Issuer: "test-issuer", SecretKey: "", ExpirationTime: 72 * time.Hour}
	_, err := NewJWT(c)
	require.EqualError(t, err, "jwt: illegal jwt configure")

	c = Config{Issuer: "test-issuer", SecretKey: "ABCDEFGH", ExpirationTime: -72 * time.Hour}
	_, err = NewJWT(c)
	require.EqualError(t, err, "jwt: illegal jwt configure")

	c = Config{Issuer: "test-issuer", SecretKey: "ABCDEFGH", ExpirationTime: 72 * time.Hour}
	j, err := NewJWT(c)
	require.NoError(t, err)
	assert.NotNil(t, j)

	tokenStr, err := j.GenToken(map[string]any{
		"user_id":   100000,
		"user_name": "test_user",
		"role_ids":  []int64{100000, 100001, 100002},
		"group":     "ADMIN",
		"is_admin":  true,
		"score":     123.123,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	payloads, err := j.ParseTokenPayloads(tokenStr)
	require.NoError(t, err)
	assert.NotEmpty(t, payloads)

	ui := &_UserInfo{}
	err = j.ParseToken(tokenStr, ui)
	require.NoError(t, err)
	assert.Equal(t, int64(100000), ui.UserID)
	assert.Equal(t, "test_user", ui.UserName)
	assert.Equal(t, []int64{100000, 100001, 100002}, ui.RoleIds)
	assert.Equal(t, "ADMIN", ui.Group)
	assert.True(t, ui.IsAdmin)
	assert.InEpsilon(t, 123.123, ui.Score, 0.0001)
}

func TestJWT_ParseToken(t *testing.T) {
	c := Config{Issuer: "test-issuer", SecretKey: "ABCDEFGH", ExpirationTime: 72 * time.Hour}
	j, err := NewJWT(c)
	require.NoError(t, err)
	assert.NotNil(t, j)

	tokenMap := map[string]any{
		"user_id":   100000,
		"user_name": "test_user",
		"role_ids":  []int64{100000, 100001, 100002},
		"group":     "ADMIN",
		"is_admin":  true,
		"score":     123.123,
	}

	tok, err := genToken("test-issuer", "ABCDEFGH", jwt.SigningMethodHS256, tokenMap, 72*time.Hour)
	require.NoError(t, err)

	uip := &_UserInfo{}
	err = j.ParseToken(tok, uip)
	require.NoError(t, err)
	assert.Equal(t, int64(100000), uip.UserID)
	assert.Equal(t, "test_user", uip.UserName)
	assert.Equal(t, []int64{100000, 100001, 100002}, uip.RoleIds)
	assert.Equal(t, "ADMIN", uip.Group)
	assert.True(t, uip.IsAdmin)
	assert.InEpsilon(t, 123.123, uip.Score, 0.0001)

	ui := _UserInfo{}
	err = j.ParseToken(tok, ui)
	require.EqualError(t, err, "decode payloads err: new map structure decoder err: result must be a pointer")

	empty := &struct{}{}
	err = j.ParseToken(tok, empty)
	require.NoError(t, err)

	another := ""
	err = j.ParseToken(tok, &another)
	require.Error(t, err)
}

func TestJWT_ParseTokenPayloads(t *testing.T) {
	c := Config{Issuer: "test-issuer", SecretKey: "ABCDEFGH", ExpirationTime: 72 * time.Hour}
	j, err := NewJWT(c)
	require.NoError(t, err)
	assert.NotNil(t, j)

	tokenMap := map[string]any{
		"user_id":   100000,
		"user_name": "test_user",
		"role_ids":  []int64{100000, 100001, 100002},
		"group":     "ADMIN",
		"is_admin":  true,
		"score":     123.123,
	}

	tok, err := genToken("another-issuer", "ABCDEFGH", jwt.SigningMethodHS256, tokenMap, 72*time.Hour)
	require.NoError(t, err)
	_, err = j.ParseTokenPayloads(tok)
	require.EqualError(t, err, "jwt.Parse err: token has invalid claims: token has invalid issuer")

	tok, err = genToken("test-issuer", "ABCD", jwt.SigningMethodHS256, tokenMap, 72*time.Hour)
	require.NoError(t, err)
	_, err = j.ParseTokenPayloads(tok)
	require.EqualError(t, err, "jwt.Parse err: token signature is invalid: signature is invalid")

	tok, err = genToken("test-issuer", "ABCDEFGH", jwt.SigningMethodHS512, tokenMap, 72*time.Hour)
	require.NoError(t, err)
	_, err = j.ParseTokenPayloads(tok)
	require.EqualError(t, err, "jwt.Parse err: token signature is invalid: signing method HS512 is invalid")

	tok, err = genToken("test-issuer", "ABCDEFGH", jwt.SigningMethodHS256, tokenMap)
	require.NoError(t, err)
	_, err = j.ParseTokenPayloads(tok)
	require.EqualError(t, err, "jwt.Parse err: token has invalid claims: token is missing required claim: exp claim is required")

	tok, err = genToken("test-issuer", "ABCDEFGH", jwt.SigningMethodHS256, tokenMap, 72*time.Hour)
	require.NoError(t, err)
	payloads, err := j.ParseTokenPayloads(tok)
	require.NoError(t, err)
	assert.NotEmpty(t, payloads)
	for k, v := range payloads {
		fmt.Printf("k: %s, v: %v, %T\n", k, v, v)
	}
}

func TestJWT_ParseTokenFromRequest(t *testing.T) {
	c := Config{Issuer: "test-issuer", SecretKey: "ABCDEFGH", ExpirationTime: 72 * time.Hour}
	j, err := NewJWT(c)
	require.NoError(t, err)
	assert.NotNil(t, j)

	tokenStr, err := j.GenToken(map[string]any{
		"user_id":   100000,
		"user_name": "test_user",
		"role_ids":  []int64{100000, 100001, 100002},
		"group":     "ADMIN",
		"is_admin":  true,
		"score":     123.123,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	_, err = j.ParseTokenPayloadsFromRequest(req)
	require.EqualError(t, err, "request.ParseFromRequest err: no token present in request")

	req.Header.Set("Authorization", "Bearer "+tokenStr)
	payloads, err := j.ParseTokenPayloadsFromRequest(req)
	require.NoError(t, err)
	assert.NotEmpty(t, payloads)
	for k, v := range payloads {
		fmt.Printf("k: %s, v: %v, %T\n", k, v, v)
	}
}

func TestJWT_ParseTokenPayloadsFromRequest(t *testing.T) {
	c := Config{Issuer: "test-issuer", SecretKey: "ABCDEFGH", ExpirationTime: 72 * time.Hour}
	j, err := NewJWT(c)
	require.NoError(t, err)
	assert.NotNil(t, j)

	tokenStr, err := j.GenToken(map[string]any{
		"user_id":   100000,
		"user_name": "test_user",
		"role_ids":  []int64{100000, 100001, 100002},
		"group":     "ADMIN",
		"is_admin":  true,
		"score":     123.123,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	uip := &_UserInfo{}
	err = j.ParseTokenFromRequest(req, uip)
	require.EqualError(t, err, "request.ParseFromRequest err: no token present in request")

	req.Header.Set("Authorization", tokenStr)
	err = j.ParseTokenFromRequest(req, uip)
	require.NoError(t, err)
	assert.Equal(t, int64(100000), uip.UserID)
	assert.Equal(t, "test_user", uip.UserName)
	assert.Equal(t, []int64{100000, 100001, 100002}, uip.RoleIds)
	assert.Equal(t, "ADMIN", uip.Group)
	assert.True(t, uip.IsAdmin)
	assert.InEpsilon(t, 123.123, uip.Score, 0.0001)
}

func genToken(issuer, secretKey string, method jwt.SigningMethod, payloads map[string]any, expirationTime ...time.Duration) (string, error) {
	claims := make(jwt.MapClaims)
	now := timex.Now()
	claims[jwtIssuer] = issuer
	claims[jwtIssueAt] = now.Unix()
	claims[jwtNotBefore] = now.Unix()
	if len(expirationTime) > 0 && expirationTime[0].Seconds() > 0 {
		claims[jwtExpire] = now.Add(expirationTime[0]).Unix()
	}

	for k, v := range payloads {
		if !slices.Contains(standardClaims, k) {
			claims[k] = v
		}
	}

	t := jwt.NewWithClaims(method, claims)
	ts, err := t.SignedString([]byte(secretKey))
	if err != nil {
		return "", errors.WithMessage(err, "sign token err")
	}

	return ts, nil
}

type _UserInfo struct {
	UserID   int64   `json:"user_id"`
	UserName string  `json:"user_name"`
	RoleIds  []int64 `json:"role_ids"`
	Group    string  `json:"group"`
	IsAdmin  bool    `json:"is_admin"`
	Score    float64 `json:"score"`
}
