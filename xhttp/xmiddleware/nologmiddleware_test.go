package xmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"github.com/sliveryou/micro-pkg/xgrpc/xinterceptor"
)

func TestNewNoLogMiddleware(t *testing.T) {
	m := NewNoLogMiddleware()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)
	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("NoLogMiddleware Handle request")
		md, ok := metadata.FromOutgoingContext(r.Context())
		assert.True(t, ok)
		data := md.Get(xinterceptor.NoLogKey)
		assert.Len(t, data, 1)
		assert.Equal(t, xinterceptor.NoLogFlag, data[0])
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
