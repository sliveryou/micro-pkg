package xreq

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMethods(t *testing.T) {
	type testCase struct {
		new func(url string, options ...Option) (*http.Request, error)
	}

	methods := map[string]testCase{
		http.MethodGet: {
			new: NewGet,
		},
		http.MethodHead: {
			new: NewHead,
		},
		http.MethodPost: {
			new: NewPost,
		},
		http.MethodPut: {
			new: NewPut,
		},
		http.MethodPatch: {
			new: NewPatch,
		},
		http.MethodDelete: {
			new: NewDelete,
		},
		http.MethodOptions: {
			new: NewOptions,
		},
	}

	for method, tc := range methods {
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			assert.Equal(t, method, r.Method)
		}))

		t.Run("new", func(t *testing.T) {
			req, err := tc.new(server.URL)
			require.NoError(t, err)

			_, err = http.DefaultClient.Do(req)
			require.NoError(t, err)
		})

		server.Close()
	}
}

func TestDoMethods(t *testing.T) {
	type testCase struct {
		do func(url string, options ...Option) (*http.Response, error)
	}

	methods := map[string]testCase{
		http.MethodGet: {
			do: Get,
		},
		http.MethodHead: {
			do: Head,
		},
		http.MethodPost: {
			do: Post,
		},
		http.MethodPut: {
			do: Put,
		},
		http.MethodPatch: {
			do: Patch,
		},
		http.MethodDelete: {
			do: Delete,
		},
		http.MethodOptions: {
			do: Options,
		},
	}

	for method, tc := range methods {
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			assert.Equal(t, method, r.Method)
		}))

		t.Run("do", func(t *testing.T) {
			_, err := tc.do(server.URL)
			require.NoError(t, err)
		})

		server.Close()
	}
}

func TestNew(t *testing.T) {
	var buffer bytes.Buffer
	ctx := context.TODO()
	request, err := New(http.MethodOptions, "http://www.test.com/api",
		Context(ctx),
		Scheme("https"),
		Host("my.test.com"),
		Path("/students"),
		BearerAuth("abcdefgh"),
		Header("Origin", "https://my.test.com"),
		Header("Access-Control-Request-Method", "PUT"),
		Header("Access-Control-Request-Headers", "Authorization", "Content-Type"),
		BodyJSON(map[string]any{"id": 1, "name": "SliverYou", "language": "go"}),
		Dump(&buffer),
	)
	require.NoError(t, err)

	header := new(bytes.Buffer)
	err = request.Header.Write(header)
	require.NoError(t, err)
	fmt.Println(header.String())

	fmt.Println(buffer.String())
}
