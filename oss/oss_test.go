package oss

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustNewOSS(t *testing.T) {
	assert.PanicsWithError(t, "oss: illegal oss cloud unknown config", func() {
		c := Config{
			Cloud: "unknown",
		}
		MustNewOSS(c)
	})
}

func TestAliyunGetURL(t *testing.T) {
	c := Config{
		NotSetACL:       true,
		Cloud:           "aliyun",
		EndPoint:        "oss-cn-hangzhou.aliyuncs.com",
		AccessKeyID:     "accessKeyID",
		AccessKeySecret: "accessKeySecret",
		BucketName:      "my-test",
	}

	o, err := NewOSS(c)
	require.NoError(t, err)
	assert.Equal(t, "https://my-test.oss-cn-hangzhou.aliyuncs.com/test/test.txt", o.GetURL("test/test.txt"))
}

func TestHuaweiGetURL(t *testing.T) {
	c := Config{
		NotSetACL:       true,
		Cloud:           "huawei",
		EndPoint:        "obs.cn-east-3.myhuaweicloud.com",
		AccessKeyID:     "accessKeyID",
		AccessKeySecret: "accessKeySecret",
		BucketName:      "my-test",
	}

	o, err := NewOSS(c)
	require.NoError(t, err)
	assert.Equal(t, "https://my-test.obs.cn-east-3.myhuaweicloud.com/test/test.txt", o.GetURL("test/test.txt"))
}

func TestTencentGetURL(t *testing.T) {
	c := Config{
		NotSetACL:       true,
		Cloud:           "tencent",
		EndPoint:        "ap-shanghai",
		AccessKeyID:     "accessKeyID",
		AccessKeySecret: "accessKeySecret",
		BucketName:      "my-test-1234567890",
	}

	o, err := NewOSS(c)
	require.NoError(t, err)
	assert.Equal(t, "https://my-test-1234567890.cos.ap-shanghai.myqcloud.com/test/test.txt", o.GetURL("test/test.txt"))
}

func TestMinIOGetURL(t *testing.T) {
	c := Config{
		NotSetACL:       true,
		Cloud:           "minio",
		EndPoint:        "localhost:9000",
		AccessKeyID:     "accessKeyID",
		AccessKeySecret: "accessKeySecret",
		BucketName:      "my-test",
	}

	o, err := NewOSS(c)
	require.NoError(t, err)
	assert.Contains(t, o.GetURL("test/test.txt"), "my-test/test/test.txt")
	fmt.Println(o.GetURL("test/test.txt"))
}

func TestLocalGetURL(t *testing.T) {
	c := Config{
		Cloud:      "local",
		EndPoint:   "my.test.com/api/bucket",
		BucketName: "testdata",
	}

	o, err := NewOSS(c)
	require.NoError(t, err)
	assert.Equal(t, "http://my.test.com/api/bucket/test/test.txt", o.GetURL("test/test.txt"))
}

func TestMockGetURL(t *testing.T) {
	c := Config{
		Cloud: "mock",
	}

	o, err := NewOSS(c)
	require.NoError(t, err)
	assert.Equal(t, "test/test.txt", o.GetURL("test/test.txt"))
}
