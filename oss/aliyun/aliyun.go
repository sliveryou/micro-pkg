package aliyun

import (
	"fmt"
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"
)

const (
	// CloudAliyun 云服务商：阿里云
	CloudAliyun = "aliyun"
)

// Option 可选配置
type Option func(o *OSS)

// WithUploadInternal 使用内网上传配置
func WithUploadInternal(uploadInternal ...bool) Option {
	return func(o *OSS) {
		o.uploadInternal = true
		if len(uploadInternal) > 0 {
			o.uploadInternal = uploadInternal[0]
		}
	}
}

// WithNotSetACL 不设置权限规则
func WithNotSetACL(notSetACL ...bool) Option {
	return func(o *OSS) {
		o.notSetACL = true
		if len(notSetACL) > 0 {
			o.notSetACL = notSetACL[0]
		}
	}
}

// OSS 阿里云 OSS 结构详情
type OSS struct {
	client           *oss.Client
	bucket           *oss.Bucket
	notSetACL        bool
	uploadInternal   bool
	externalEndpoint string
}

// NewOSS 创建一个阿里云 OSS 对象
func NewOSS(endpoint, accessKeyID, accessKeySecret, bucketName string, opts ...Option) (*OSS, error) {
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, errors.WithMessage(err, "aliyun: new oss err")
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, errors.WithMessage(err, "aliyun: oss get bucket err")
	}

	o := &OSS{client: client, bucket: bucket, externalEndpoint: endpoint}
	for _, opt := range opts {
		opt(o)
	}
	if !o.notSetACL {
		// 设置存储桶公共读权限规则
		err = client.SetBucketACL(bucketName, oss.ACLPublicRead)
		if err != nil {
			return nil, errors.WithMessage(err, "aliyun: oss set bucket acl err")
		}
	}
	if o.uploadInternal {
		o.externalEndpoint = strings.Replace(o.externalEndpoint, "-internal", "", 1)
	}

	return o, nil
}

// Cloud 获取云服务商名称
func (o *OSS) Cloud() string {
	return CloudAliyun
}

// GetURL 获取对象在阿里云 OSS 上的完整访问 URL
func (o *OSS) GetURL(key string) string {
	if strings.HasPrefix(key, "http") {
		return key
	}

	return fmt.Sprintf("https://%s.%s/%s", o.bucket.BucketName, o.externalEndpoint, key)
}

// GetObject 获取对象在阿里云 OSS 的存储数据
func (o *OSS) GetObject(key string) (io.ReadCloser, error) {
	obj, err := o.bucket.GetObject(key)
	if err != nil {
		return nil, errors.WithMessage(err, "aliyun: oss get object err")
	}

	return obj, nil
}

// PutObject 上传对象至阿里云 OSS
func (o *OSS) PutObject(key string, reader io.Reader) (string, error) {
	err := o.bucket.PutObject(key, reader)
	if err != nil {
		return "", errors.WithMessage(err, "aliyun: oss put object err")
	}

	return o.GetURL(key), nil
}

// DeleteObjects 批量删除阿里云 OSS 上的对象
func (o *OSS) DeleteObjects(keys ...string) error {
	_, err := o.bucket.DeleteObjects(keys)

	return errors.WithMessage(err, "aliyun: oss delete objects err")
}

// UploadFile 上传文件至阿里云 OSS，filePath：文件路径，partSize：分块大小（字节），routines：并发数
func (o *OSS) UploadFile(key, filePath string, partSize int64, routines int) (string, error) {
	err := o.bucket.UploadFile(key, filePath, partSize,
		oss.Routines(routines),
		oss.Checkpoint(true, ""))
	if err != nil {
		return "", errors.WithMessage(err, "aliyun: oss upload file err")
	}

	return o.GetURL(key), nil
}

// AuthorizedUpload 授权上传至阿里云 OSS，expires：过期时间（秒）
func (o *OSS) AuthorizedUpload(key string, expires int) (string, error) {
	signedURL, err := o.bucket.SignURL(key, oss.HTTPPut, int64(expires))
	if err != nil {
		return "", errors.WithMessage(err, "aliyun: oss authorized upload err")
	}

	return signedURL, nil
}

// GetThumbnailSuffix 获取缩略图后缀
func (o *OSS) GetThumbnailSuffix(width, height int, size int64) string {
	// 参考文档 https://help.aliyun.com/document_detail/44688.html
	var suffix string
	maxSize := int64(20 << 20) // 20M
	if (width > 0 || height > 0) && size <= maxSize {
		max := 4096
		suffix = "?x-oss-process=image/resize,m_fixed,limit_0"
		if width > 0 && width <= max {
			suffix += fmt.Sprintf(",w_%d", width)
		}

		if height > 0 && height <= max {
			suffix += fmt.Sprintf(",h_%d", height)
		}
	}

	return suffix
}
