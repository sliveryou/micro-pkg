package tencent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	endpoint        = "ap-shanghai"
	accessKeyID     = "accessKeyID"
	accessKeySecret = "accessKeySecret"
	bucketName      = "my-test-1234567890"
)

func TestNewCOS(t *testing.T) {
	cos, err := NewCOS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	assert.NotNil(t, cos)
}

func TestCOS_Cloud(t *testing.T) {
	cos, err := NewCOS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	assert.Equal(t, "tencent", cos.Cloud())
}

func TestCOS_GetURL(t *testing.T) {
	cos, err := NewCOS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	t.Log(cos.GetURL("test/test.txt"))
}

func TestCOS_AuthorizedUpload(t *testing.T) {
	cos, err := NewCOS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	t.Log(cos.AuthorizedUpload("test/test.txt", 120))
}
