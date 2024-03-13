package xmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCorsMiddleware_Options(t *testing.T) {
	m := NewCorsMiddleware()
	req := httptest.NewRequest(http.MethodOptions, "http://localhost", http.NoBody)
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("CorsMiddleware Handle request")
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNoContent, resp.Code)
	t.Log(resp.Header())
	assert.Equal(t, "Content-Type, Authorization, AccessToken, Token, X-CSRF-Token, X-Health-Secret", resp.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", resp.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Length, Content-Type, Access-Control-Allow-Origin, Access-Control-Allow-Headers, X-Ca-Error-Code, X-Ca-Error-Message", resp.Header().Get("Access-Control-Expose-Headers"))
}

func TestNewCorsMiddleware_Get(t *testing.T) {
	m := NewCorsMiddleware()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("CorsMiddleware Handle request")
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	t.Log(resp.Header())
	assert.Equal(t, "Content-Type, Authorization, AccessToken, Token, X-CSRF-Token, X-Health-Secret", resp.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", resp.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Length, Content-Type, Access-Control-Allow-Origin, Access-Control-Allow-Headers, X-Ca-Error-Code, X-Ca-Error-Message", resp.Header().Get("Access-Control-Expose-Headers"))
}
