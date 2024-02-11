package tencent

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cossdk "github.com/tencentyun/cos-go-sdk-v5"

	"github.com/sliveryou/micro-pkg/xhttp"
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

func TestCloneObjectPutOptions(t *testing.T) {
	key := "test/test.txt"
	opt := &cossdk.ObjectPutOptions{
		ObjectPutHeaderOptions: &cossdk.ObjectPutHeaderOptions{
			ContentType:   xhttp.TypeByExtension(key),
			ContentLength: 1024,
			Listener:      &cossdk.DefaultProgressListener{},
		},
	}
	cloneOpt := cossdk.CloneObjectPutOptions(opt)
	fmt.Printf("%+v\n", cloneOpt.ObjectPutHeaderOptions)
}
