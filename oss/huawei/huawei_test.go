package huawei

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	endpoint        = "obs.cn-east-3.myhuaweicloud.com"
	accessKeyID     = "accessKeyID"
	accessKeySecret = "accessKeySecret"
	bucketName      = "my-test"
)

func TestNewOBS(t *testing.T) {
	obs, err := NewOBS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	assert.NotNil(t, obs)
}

func TestOBS_Cloud(t *testing.T) {
	obs, err := NewOBS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	assert.Equal(t, "huawei", obs.Cloud())
}

func TestOBS_GetURL(t *testing.T) {
	obs, err := NewOBS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	t.Log(obs.GetURL("test/test.txt"))
}

func TestOBS_AuthorizedUpload(t *testing.T) {
	obs, err := NewOBS(endpoint, accessKeyID, accessKeySecret, bucketName, WithNotSetACL())
	require.NoError(t, err)
	t.Log(obs.AuthorizedUpload("test/test.txt", 120))
}
