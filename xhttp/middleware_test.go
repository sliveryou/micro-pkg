package xhttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"github.com/sliveryou/micro-pkg/errcode"
)

func TestNewCorsMiddleware(t *testing.T) {
	m := NewCorsMiddleware()
	req := httptest.NewRequest(http.MethodOptions, "http://localhost", nil)
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
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
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

		w.Header().Set(HeaderContentType, ApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":200,"msg":"ok"}`))
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestNewIgnoreRLogMiddleware(t *testing.T) {
	m := NewIgnoreRLogMiddleware()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
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
