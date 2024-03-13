package xmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sliveryou/micro-pkg/errcode"
)

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
