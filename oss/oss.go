package oss

import (
	"io"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/oss/aliyun"
	"github.com/sliveryou/micro-pkg/oss/huawei"
	"github.com/sliveryou/micro-pkg/oss/local"
	"github.com/sliveryou/micro-pkg/oss/minio"
	"github.com/sliveryou/micro-pkg/oss/mock"
	"github.com/sliveryou/micro-pkg/oss/tencent"
	"github.com/sliveryou/micro-pkg/xhttp"
)

// OSS 客户端接口
type OSS interface {
	// Cloud 获取云服务商名称
	Cloud() string
	// GetURL 获取对象在 OSS 上的完整访问 URL
	GetURL(key string) string
	// GetObject 获取对象在 OSS 的存储数据
	GetObject(key string) (io.ReadCloser, error)
	// PutObject 上传对象至 OSS
	PutObject(key string, reader io.Reader) (string, error)
	// DeleteObjects 批量删除 OSS 上的对象
	DeleteObjects(keys ...string) error
	// UploadFile 上传文件至 OSS，filePath：文件路径，partSize：分块大小（字节），routines：并发数
	UploadFile(key, filePath string, partSize int64, routines int) (string, error)
	// AuthorizedUpload 授权上传至 OSS，expires：过期时间（秒）
	AuthorizedUpload(key string, expires int) (string, error)
	// GetThumbnailSuffix 获取缩略图后缀，如果只传一个值则进行等比缩放，两个值都传时会强制缩放，可能会导致图片变形
	GetThumbnailSuffix(width, height int, size int64) string
}

// Config OSS 配置
type Config struct {
	UseSSL          bool   `json:",optional"`                                         // 是否使用安全配置（用于 minio 和 local 云服务商模式）
	UploadInternal  bool   `json:",optional"`                                         // 是否使用内网上传（用于 aliyun 云服务商模式）
	NotSetACL       bool   `json:",optional"`                                         // 不设置权限规则
	Cloud           string `json:",options=[aliyun,huawei,tencent,minio,local,mock]"` // 云服务商（当前支持 aliyun、huawei、tencent、minio、local 和 mock）
	EndPoint        string `json:",optional"`                                         // 端节点
	AccessKeyID     string `json:",optional"`                                         // 访问鉴权ID
	AccessKeySecret string `json:",optional"`                                         // 访问鉴权私钥
	BucketName      string `json:",optional"`                                         // 存储桶名称
}

// checkConfig 检查配置
func checkConfig(c Config) (err error) {
	switch cloud := c.Cloud; cloud {
	case mock.CloudMock:
	case local.CloudLocal:
		if c.BucketName == "" {
			err = errors.New("oss: illegal oss cloud local config")
		}
	default:
		if c.EndPoint == "" || c.AccessKeyID == "" ||
			c.AccessKeySecret == "" || c.BucketName == "" {
			err = errors.Errorf("oss: illegal oss cloud %s config", cloud)
		}
	}

	return
}

// NewOSS 新建 OSS 客户端
func NewOSS(c Config) (OSS, error) {
	if err := checkConfig(c); err != nil {
		return nil, err
	}

	var client OSS
	var err error
	parsedEndpoint, useSSL := xhttp.ParseEndpoint(c.EndPoint)

	switch c.Cloud {
	case aliyun.CloudAliyun:
		client, err = aliyun.NewOSS(parsedEndpoint, c.AccessKeyID, c.AccessKeySecret, c.BucketName,
			aliyun.WithUploadInternal(c.UploadInternal),
			aliyun.WithNotSetACL(c.NotSetACL))
	case huawei.CloudHuawei:
		client, err = huawei.NewOBS(parsedEndpoint, c.AccessKeyID, c.AccessKeySecret, c.BucketName,
			huawei.WithNotSetACL(c.NotSetACL))
	case tencent.CloudTencent:
		client, err = tencent.NewCOS(parsedEndpoint, c.AccessKeyID, c.AccessKeySecret, c.BucketName,
			tencent.WithNotSetACL(c.NotSetACL))
	case minio.CloudMinIO:
		client, err = minio.NewMinIO(parsedEndpoint, c.AccessKeyID, c.AccessKeySecret, c.BucketName,
			minio.WithSecure(c.UseSSL || useSSL),
			minio.WithNotSetACL(c.NotSetACL))
	case local.CloudLocal:
		client, err = local.NewLSS(parsedEndpoint, c.BucketName,
			local.WithSecure(c.UseSSL || useSSL))
	default:
		client = mock.NewMSS()
	}
	if err != nil {
		return nil, errors.WithMessage(err, "oss: new oss client err")
	}

	return &defaultOSS{c: c, client: client}, nil
}

// MustNewOSS 新建 OSS 客户端
func MustNewOSS(c Config) OSS {
	o, err := NewOSS(c)
	if err != nil {
		panic(err)
	}

	return o
}

// defaultOSS 默认 OSS 客户端
type defaultOSS struct {
	c      Config
	client OSS
}

// Cloud 获取云服务商名称
func (o *defaultOSS) Cloud() string {
	return o.client.Cloud()
}

// GetURL 获取对象在 OSS 上的完整访问 URL
func (o *defaultOSS) GetURL(key string) string {
	return o.client.GetURL(key)
}

// GetObject 获取对象在 OSS 的存储数据
func (o *defaultOSS) GetObject(key string) (io.ReadCloser, error) {
	return o.client.GetObject(key)
}

// PutObject 上传对象至 OSS
func (o *defaultOSS) PutObject(key string, reader io.Reader) (string, error) {
	return o.client.PutObject(key, reader)
}

// DeleteObjects 批量删除 OSS 上的对象
func (o *defaultOSS) DeleteObjects(keys ...string) error {
	return o.client.DeleteObjects(keys...)
}

// UploadFile 上传文件至 OSS，filePath：文件路径，partSize：分块大小（字节），routines：并发数
func (o *defaultOSS) UploadFile(key, filePath string, partSize int64, routines int) (string, error) {
	// 默认分片大小为 5MB
	if partSize <= 0 {
		partSize = 5 * 1024 * 1024
	}
	// 默认上传并发数为 5
	if routines <= 0 {
		routines = 5
	}

	return o.client.UploadFile(key, filePath, partSize, routines)
}

// AuthorizedUpload 授权上传至 OSS，expires：过期时间（秒）
func (o *defaultOSS) AuthorizedUpload(key string, expires int) (string, error) {
	// 默认过期时间为120s
	if expires <= 0 {
		expires = 120
	}

	return o.client.AuthorizedUpload(key, expires)
}

// GetThumbnailSuffix 获取缩略图后缀
func (o *defaultOSS) GetThumbnailSuffix(width, height int, size int64) string {
	return o.client.GetThumbnailSuffix(width, height, size)
}
