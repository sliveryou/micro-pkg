package mock

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMSS(t *testing.T) {
	mss := NewMSS()

	out, err := mss.PutObject("test/test.txt", strings.NewReader("test-oss"))
	require.NoError(t, err)
	t.Log(out, mss.GetURL(out))

	err = mss.DeleteObjects("test/test.txt")
	require.NoError(t, err)
}

func TestLSS_Cloud(t *testing.T) {
	mss := NewMSS()
	assert.Equal(t, "mock", mss.Cloud())
}

func TestLSS_GetURL(t *testing.T) {
	mss := NewMSS()
	t.Log(mss.GetURL("test/test.txt"))
}

func TestLSS_AuthorizedUpload(t *testing.T) {
	mss := NewMSS()
	t.Log(mss.AuthorizedUpload("test/test.txt", 120))
}
