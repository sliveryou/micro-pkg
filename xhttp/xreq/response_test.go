package xreq

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
	defer server.Close()

	client := NewClient(
		URL(server.URL),
		Path("api", "/students"),
		BearerAuth("abcdefgh"),
	)

	response, err := client.Post(BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}))
	require.NoError(t, err)

	result := make(map[string]string)
	err = response.Unmarshal(&result)
	require.NoError(t, err)

	assert.True(t, response.IsSuccess())
	assert.False(t, response.IsError())
	assert.Equal(t, "application/json; charset=utf-8", response.ContentType())
	assert.Equal(t, int64(36), response.Size())
	assert.Equal(t, `{"language":"go","name":"SliverYou"}`, response.String())
	assert.Equal(t, map[string]string{"name": "SliverYou", "language": "go"}, result)
	assert.Contains(t, response.GetAllHeadersString(), "Content-Length: 36\r\nContent-Type: application/json; charset=utf-8\r\nDate:")
	t.Log(response.ReceivedAt())
}
