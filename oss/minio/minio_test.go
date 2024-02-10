package minio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	endpoint        = "localhost:9000"
	accessKeyID     = "minio"
	accessKeySecret = "minio123456"
	bucketName      = "my-test"
)

func getMio() (*MinIO, error) {
	return NewMinIO(endpoint, accessKeyID, accessKeySecret, bucketName,
		WithSecure(false), WithNotSetACL())
}

func TestNewMinIO(t *testing.T) {
	mio, err := getMio()
	require.NoError(t, err)
	assert.NotNil(t, mio)
}

func TestMinIO_Cloud(t *testing.T) {
	mio, err := getMio()
	require.NoError(t, err)
	assert.Equal(t, "minio", mio.Cloud())
}

func TestMinIO_GetURL(t *testing.T) {
	mio, err := getMio()
	require.NoError(t, err)
	t.Log(mio.GetURL("test/test.txt"))
}

func TestMinIO_AuthorizedUpload(t *testing.T) {
	mio, err := getMio()
	require.NoError(t, err)
	t.Log(mio.AuthorizedUpload("test/test.txt", 120))
}
