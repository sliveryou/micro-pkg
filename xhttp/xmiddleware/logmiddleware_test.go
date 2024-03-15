package xmiddleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sliveryou/micro-pkg/xhttp"
)

func TestNewLogMiddleware(t *testing.T) {
	m := NewLogMiddleware()
	req := httptest.NewRequest(http.MethodPost, "http://localhost", strings.NewReader(`{"id":1}`))
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("LogMiddleware Handle request")

		var buf bytes.Buffer
		io.Copy(&buf, req.Body)
		t.Log(buf.String())

		w.Header().Set(xhttp.HeaderContentType, xhttp.MIMEApplicationJSON)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":200,"msg":"ok"}`))
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
