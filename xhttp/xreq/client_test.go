package xreq

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/xhttp"
)

func TestClient_Do(t *testing.T) {
	_, err := NewClient().Do("unknown method")
	require.EqualError(t, err, `new http request err: net/http: invalid method "unknown method"`)
}

func TestClientDoMethods(t *testing.T) {
	type testCase struct {
		clientDo func(options ...Option) (*Response, error)
	}

	client := DefaultHTTPClient
	methods := map[string]testCase{
		http.MethodGet: {
			clientDo: NewClientWithHTTPClient(client).Get,
		},
		http.MethodHead: {
			clientDo: NewClientWithHTTPClient(client).Head,
		},
		http.MethodPost: {
			clientDo: NewClientWithHTTPClient(client).Post,
		},
		http.MethodPut: {
			clientDo: NewClientWithHTTPClient(client).Put,
		},
		http.MethodPatch: {
			clientDo: NewClientWithHTTPClient(client).Patch,
		},
		http.MethodDelete: {
			clientDo: NewClientWithHTTPClient(client).Delete,
		},
		http.MethodOptions: {
			clientDo: NewClientWithHTTPClient(client).Options,
		},
	}

	for method, tc := range methods {
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			assert.Equal(t, method, r.Method)
		}))

		t.Run("clientDo", func(t *testing.T) {
			_, err := tc.clientDo(URL(server.URL))
			require.NoError(t, err)
		})

		server.Close()
	}
}

func TestClient(t *testing.T) {
	type args struct {
		method string
		url    string
		header map[string]string
		body   io.Reader
	}
	cases := []struct {
		title string
		args  args
	}{
		{
			title: "do get ip",
			args: args{
				method: "GET",
				url:    "https://www.httpbin.org/ip",
				header: map[string]string{
					xhttp.HeaderAcceptLanguage: "zh-CN,zh;q=0.9,en;q=0.8",
					xhttp.HeaderUserAgent:      "Go-HTTP-Request",
				},
				body: nil,
			},
		},
		{
			title: "do get method",
			args: args{
				method: "GET",
				url:    "https://www.httpbin.org/get",
				header: map[string]string{
					xhttp.HeaderAcceptLanguage: "zh-CN,zh;q=0.9,en;q=0.8",
					xhttp.HeaderUserAgent:      "Go-HTTP-Request",
				},
				body: nil,
			},
		},
		{
			title: "do post method",
			args: args{
				method: "POST",
				url:    "https://www.httpbin.org/post",
				header: map[string]string{
					xhttp.HeaderAcceptLanguage: "zh-CN,zh;q=0.9,en;q=0.8",
					xhttp.HeaderUserAgent:      "Go-HTTP-Request",
				},
				body: strings.NewReader("a=b&c=d"),
			},
		},
	}

	client := NewClient()
	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			resp, err := client.Do(c.args.method,
				URL(c.args.url), HeaderMap(c.args.header), BodyReader(c.args.body))
			if err == nil {
				t.Log(resp.String())
			}
		})
	}
}
