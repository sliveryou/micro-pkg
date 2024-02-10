package aliyun

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	endpoint        = "oss-cn-hangzhou.aliyuncs.com"
	accessKeyID     = "accessKeyID"
	accessKeySecret = "accessKeySecret"
	bucketName      = "my-test"
)

func TestNewOSS(t *testing.T) {
	oss, err := NewOSS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	assert.NotNil(t, oss)
}

func TestOSS_Cloud(t *testing.T) {
	oss, err := NewOSS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	assert.Equal(t, "aliyun", oss.Cloud())
}

func TestOSS_GetURL(t *testing.T) {
	oss, err := NewOSS("oss-cn-hangzhou-internal.aliyuncs.com", accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	assert.Equal(t, "https://my-test.oss-cn-hangzhou-internal.aliyuncs.com/test/test.txt", oss.GetURL("test/test.txt"))

	oss, err = NewOSS("oss-cn-hangzhou-internal.aliyuncs.com", accessKeyID, accessKeySecret, bucketName, WithNotSetACL(), WithUploadInternal(true))
	require.NoError(t, err)
	assert.Equal(t, "https://my-test.oss-cn-hangzhou.aliyuncs.com/test/test.txt", oss.GetURL("test/test.txt"))
}

func TestOSS_AuthorizedUpload(t *testing.T) {
	oss, err := NewOSS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	t.Log(oss.AuthorizedUpload("test/test.txt", 120))
}
