package xmiddleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/xhttp"
)

func TestCORSMiddleware_Handle_Options(t *testing.T) {
	m := NewCORSMiddleware()

	req := httptest.NewRequest(http.MethodOptions, "http://localhost", http.NoBody)
	req.Header.Set(xhttp.HeaderOrigin, "http://my.test.com")
	req.Header.Set(xhttp.HeaderAccessControlRequestMethod, xhttp.MethodPost)
	req.Header.Add(xhttp.HeaderAccessControlRequestHeaders, xhttp.HeaderAuthorization)
	req.Header.Add(xhttp.HeaderAccessControlRequestHeaders, strings.Join([]string{xhttp.HeaderContentType, xhttp.HeaderXRequestedWith}, sep))

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("CORSMiddleware Handle request")
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Equal(t, "*", resp.Header().Get(xhttp.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "Authorization, Content-Type, X-Requested-With", resp.Header().Get(xhttp.HeaderAccessControlAllowHeaders))
	assert.Equal(t, "POST", resp.Header().Get(xhttp.HeaderAccessControlAllowMethods))
	assert.Equal(t, []string{xhttp.HeaderOrigin, xhttp.HeaderAccessControlRequestHeaders, xhttp.HeaderAccessControlRequestMethod}, resp.Header()[xhttp.HeaderVary])

	dump, err := httputil.DumpResponse(resp.Result(), true)
	require.NoError(t, err)
	fmt.Println(string(dump))
}

func TestCORSMiddleware_Handle_Options_MethodNotAllowed(t *testing.T) {
	m := NewCORSMiddleware()

	req := httptest.NewRequest(http.MethodOptions, "http://localhost", http.NoBody)
	req.Header.Set(xhttp.HeaderOrigin, "http://my.test.com")
	req.Header.Set(xhttp.HeaderAccessControlRequestMethod, xhttp.MethodTrace)
	req.Header.Add(xhttp.HeaderAccessControlRequestHeaders, xhttp.HeaderAuthorization)
	req.Header.Add(xhttp.HeaderAccessControlRequestHeaders, strings.Join([]string{xhttp.HeaderContentType, xhttp.HeaderXRequestedWith}, sep))

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("CORSMiddleware Handle request")
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Empty(t, resp.Header().Get(xhttp.HeaderAccessControlAllowOrigin))
	assert.Empty(t, resp.Header().Get(xhttp.HeaderAccessControlAllowHeaders))
	assert.Empty(t, resp.Header().Get(xhttp.HeaderAccessControlAllowMethods))
	assert.Equal(t, []string{xhttp.HeaderOrigin}, resp.Header()[xhttp.HeaderVary])

	dump, err := httputil.DumpResponse(resp.Result(), true)
	require.NoError(t, err)
	fmt.Println(string(dump))
}

func TestCORSMiddleware_Handle_Post(t *testing.T) {
	m := NewCORSMiddleware()

	req := httptest.NewRequest(http.MethodPost, "http://localhost", strings.NewReader(`{"a":"a","b":1}`))
	req.Header.Set(xhttp.HeaderOrigin, "http://my.test.com")
	req.Header.Set(xhttp.HeaderContentType, xhttp.MIMEApplicationJSON)

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("CORSMiddleware Handle request")
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "*", resp.Header().Get(xhttp.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "Content-Disposition, Content-Encoding, X-Ca-Error-Code, X-Ca-Error-Message", resp.Header().Get(xhttp.HeaderAccessControlExposeHeaders))
	assert.Equal(t, []string{xhttp.HeaderOrigin}, resp.Header()[xhttp.HeaderVary])

	dump, err := httputil.DumpResponse(resp.Result(), true)
	require.NoError(t, err)
	fmt.Println(string(dump))
}

func TestCORSMiddleware_Handle_Trace_Fail(t *testing.T) {
	m := NewCORSMiddleware()

	req := httptest.NewRequest(http.MethodTrace, "http://localhost", strings.NewReader(`{"a":"a","b":1}`))
	req.Header.Set(xhttp.HeaderOrigin, "http://my.test.com")
	req.Header.Set(xhttp.HeaderContentType, xhttp.MIMEApplicationJSON)

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("CORSMiddleware Handle request")
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, resp.Header().Get(xhttp.HeaderAccessControlAllowOrigin))
	assert.Empty(t, resp.Header().Get(xhttp.HeaderAccessControlExposeHeaders))
	assert.Equal(t, []string{xhttp.HeaderOrigin}, resp.Header()[xhttp.HeaderVary])

	dump, err := httputil.DumpResponse(resp.Result(), true)
	require.NoError(t, err)
	fmt.Println(string(dump))
}

func TestAllowAllCORSMiddleware_Handle_Trace_OK(t *testing.T) {
	m := AllowAllCORSMiddleware()

	req := httptest.NewRequest(http.MethodTrace, "http://localhost", strings.NewReader(`{"a":"a","b":1}`))
	req.Header.Set(xhttp.HeaderOrigin, "http://my.test.com")
	req.Header.Set(xhttp.HeaderContentType, xhttp.MIMEApplicationJSON)

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("CORSMiddleware Handle request")
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "*", resp.Header().Get(xhttp.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "Content-Disposition, Content-Encoding, X-Ca-Error-Code, X-Ca-Error-Message", resp.Header().Get(xhttp.HeaderAccessControlExposeHeaders))
	assert.Equal(t, []string{xhttp.HeaderOrigin}, resp.Header()[xhttp.HeaderVary])

	dump, err := httputil.DumpResponse(resp.Result(), true)
	require.NoError(t, err)
	fmt.Println(string(dump))
}

func TestUnsafeAllowAllCORSMiddleware_Handle_Options(t *testing.T) {
	m := UnsafeAllowAllCORSMiddleware()

	req := httptest.NewRequest(http.MethodOptions, "http://localhost", http.NoBody)
	req.Header.Set(xhttp.HeaderOrigin, "http://my.test.com")
	req.Header.Set(xhttp.HeaderAccessControlRequestMethod, xhttp.MethodPost)
	req.Header.Add(xhttp.HeaderAccessControlRequestHeaders, xhttp.HeaderAuthorization)
	req.Header.Add(xhttp.HeaderAccessControlRequestHeaders, strings.Join([]string{xhttp.HeaderContentType, xhttp.HeaderXRequestedWith}, sep))

	handler := m.Handle(func(w http.ResponseWriter, r *http.Request) {
		t.Log("CORSMiddleware Handle request")
	})

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNoContent, resp.Code)
	assert.Equal(t, "http://my.test.com", resp.Header().Get(xhttp.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "Authorization, Content-Type, X-Requested-With", resp.Header().Get(xhttp.HeaderAccessControlAllowHeaders))
	assert.Equal(t, "true", resp.Header().Get(xhttp.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "POST", resp.Header().Get(xhttp.HeaderAccessControlAllowMethods))
	assert.Equal(t, []string{xhttp.HeaderOrigin, xhttp.HeaderAccessControlRequestHeaders, xhttp.HeaderAccessControlRequestMethod}, resp.Header()[xhttp.HeaderVary])

	dump, err := httputil.DumpResponse(resp.Result(), true)
	require.NoError(t, err)
	fmt.Println(string(dump))
}

func TestCORSMiddleware_isOriginAllowed(t *testing.T) {
	c := DefaultCORSConfig()
	c.AllowOrigins = []string{"https://*.test.com"}
	m := NewCORSMiddleware(c)

	assert.True(t, m.isOriginAllowed("https://my.test.com"))
	assert.True(t, m.isOriginAllowed("https://abc.test.com"))
	assert.False(t, m.isOriginAllowed("http://my.test.com"))
	assert.False(t, m.isOriginAllowed("http://abc.test.com"))
	assert.False(t, m.isOriginAllowed("https://abc.xtest.com"))
}
