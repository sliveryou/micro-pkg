package local

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	endpoint   = "endpoint"
	bucketName = "testdata"
)

func TestNewLSS(t *testing.T) {
	lss, err := NewLSS(endpoint, bucketName)
	require.NoError(t, err)
	defer os.RemoveAll(bucketName)

	out, err := lss.PutObject("test/test.txt", strings.NewReader("test-oss"))
	require.NoError(t, err)
	t.Log(out, lss.GetURL(out))

	err = lss.DeleteObjects("test/test.txt")
	require.NoError(t, err)
}

func TestLSS_Cloud(t *testing.T) {
	lss, err := NewLSS(endpoint, bucketName)
	require.NoError(t, err)
	assert.Equal(t, "local", lss.Cloud())
}

func TestLSS_GetURL(t *testing.T) {
	lss, err := NewLSS(endpoint, bucketName)
	require.NoError(t, err)
	t.Log(lss.GetURL("test/test.txt"))
}

func TestLSS_AuthorizedUpload(t *testing.T) {
	lss, err := NewLSS(endpoint, bucketName)
	require.NoError(t, err)
	t.Log(lss.AuthorizedUpload("test/test.txt", 120))
}
