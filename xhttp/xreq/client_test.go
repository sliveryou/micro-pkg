package xreq

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Do(t *testing.T) {
	_, err := NewClient(http.DefaultClient).Do("unknown method")
	require.EqualError(t, err, `new http request err: net/http: invalid method "unknown method"`)
}

func TestClientDoMethods(t *testing.T) {
	type testCase struct {
		clientDo func(options ...Option) (*http.Response, error)
	}

	client := http.DefaultClient
	methods := map[string]testCase{
		http.MethodGet: {
			clientDo: NewClient(client).Get,
		},
		http.MethodHead: {
			clientDo: NewClient(client).Head,
		},
		http.MethodPost: {
			clientDo: NewClient(client).Post,
		},
		http.MethodPut: {
			clientDo: NewClient(client).Put,
		},
		http.MethodPatch: {
			clientDo: NewClient(client).Patch,
		},
		http.MethodDelete: {
			clientDo: NewClient(client).Delete,
		},
		http.MethodOptions: {
			clientDo: NewClient(client).Options,
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
