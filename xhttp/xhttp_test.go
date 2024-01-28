package xhttp

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueries(t *testing.T) {
	r, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://test.com/api/test?id=1&hash=a&hash=b", nil)
	require.NoError(t, err)

	assert.Equal(t, "1", Query(r, "id"))
	assert.Equal(t, []string{"1"}, QueryArray(r, "id"))
	assert.Equal(t, "a", Query(r, "hash"))
	assert.Equal(t, []string{"a", "b"}, QueryArray(r, "hash"))
}

func TestGetInternalIP(t *testing.T) {
	iip := GetInternalIP()
	assert.NotEmpty(t, iip)
	t.Log(iip)
}
