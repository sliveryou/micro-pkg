package tencent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	cos "github.com/tencentyun/cos-go-sdk-v5"

	"github.com/sliveryou/micro-pkg/xhttp"
)

const (
	// CloudTencent 云服务商：腾讯云
	CloudTencent = "tencent"
)

// OptionFunc 可选配置
type OptionFunc func(c *COS)

// WithNotSetACL 不设置权限规则
func WithNotSetACL(notSetACL ...bool) OptionFunc {
	return func(c *COS) {
		c.notSetACL = true
		if len(notSetACL) > 0 {
			c.notSetACL = notSetACL[0]
		}
	}
}

// COS 腾讯云 COS 结构详情
type COS struct {
	notSetACL bool
	ak        string
	sk        string
	client    *cos.Client
}

// NewCOS 创建一个腾讯云 COS 对象
func NewCOS(endpoint, accessKeyID, accessKeySecret, bucketName string, opts ...OptionFunc) (*COS, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucketName, endpoint))
	if err != nil {
		return nil, errors.WithMessage(err, "tencent: new cos err")
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessKeyID,
			SecretKey: accessKeySecret,
		},
	})

	c := &COS{ak: accessKeyID, sk: accessKeySecret, client: client}
	for _, opt := range opts {
		opt(c)
	}
	if !c.notSetACL {
		// 设置存储桶公共读权限规则
		_, err = client.Bucket.PutACL(context.Background(), &cos.BucketPutACLOptions{
			Header: &cos.ACLHeaderOptions{XCosACL: "public-read"},
		})
		if err != nil {
			return nil, errors.WithMessage(err, "tencent: cos set bucket acl err")
		}
	}

	return c, nil
}

// Cloud 获取云服务商名称
func (c *COS) Cloud() string {
	return CloudTencent
}

// GetURL 获取对象在腾讯云 COS 上的完整访问 URL
func (c *COS) GetURL(key string) string {
	if strings.HasPrefix(key, "http") {
		return key
	}

	return fmt.Sprintf("%s/%s", c.client.BaseURL.BucketURL, key)
}

// GetObject 获取对象在腾讯云 COS 的存储数据
func (c *COS) GetObject(key string) (io.ReadCloser, error) {
	obj, err := c.client.Object.Get(context.Background(), key, &cos.ObjectGetOptions{})
	if err != nil {
		return nil, errors.WithMessage(err, "tencent: cos get object err")
	}

	return obj.Body, nil
}

// PutObject 上传对象至腾讯云 COS
func (c *COS) PutObject(key string, reader io.Reader) (string, error) {
	contentLength, _ := xhttp.GetReaderLen(reader)

	_, err := c.client.Object.Put(context.Background(), key, reader, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType:   xhttp.TypeByExtension(key),
			ContentLength: contentLength,
			Listener:      &cos.DefaultProgressListener{},
		},
	})
	if err != nil {
		return "", errors.WithMessage(err, "tencent: cos put object err")
	}

	return c.GetURL(key), nil
}

// DeleteObjects 批量删除腾讯云 COS 上的对象
func (c *COS) DeleteObjects(keys ...string) error {
	objects := make([]cos.Object, 0, len(keys))
	for _, key := range keys {
		objects = append(objects, cos.Object{Key: key})
	}

	option := &cos.ObjectDeleteMultiOptions{Objects: objects, Quiet: true}
	_, _, err := c.client.Object.DeleteMulti(context.Background(), option)

	return errors.WithMessage(err, "tencent: cos delete objects err")
}

// UploadFile 上传文件至腾讯云 COS，filePath：文件路径，partSize：分块大小（字节），routines：并发数
func (c *COS) UploadFile(key, filePath string, partSize int64, routines int) (string, error) {
	_, _, err := c.client.Object.Upload(context.Background(), key, filePath, &cos.MultiUploadOptions{
		PartSize:       partSize / 1024 / 1024,
		ThreadPoolSize: routines,
		CheckPoint:     true,
		OptIni: &cos.InitiateMultipartUploadOptions{
			ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
				ContentType: xhttp.TypeByExtension(filePath),
			},
		},
	})
	if err != nil {
		return "", errors.WithMessage(err, "tencent: cos upload file err")
	}

	return c.GetURL(key), nil
}

// AuthorizedUpload 授权上传至腾讯云 COS，expires：过期时间（秒）
func (c *COS) AuthorizedUpload(key string, expires int) (string, error) {
	signedURL, err := c.client.Object.GetPresignedURL(context.Background(), http.MethodPut, key,
		c.ak, c.sk, time.Duration(expires)*time.Second, nil)
	if err != nil {
		return "", errors.WithMessage(err, "tencent: cos authorized upload err")
	}

	return signedURL.String(), nil
}

// GetThumbnailSuffix 获取缩略图后缀
func (c *COS) GetThumbnailSuffix(width, height int, size int64) string {
	// 参考文档 https://cloud.tencent.com/document/product/436/44880
	var suffix string
	maxSize := int64(32 << 2) // 32M
	if (width > 0 || height > 0) && size <= maxSize {
		suffix = "?imageMogr2/thumbnail/"
		max := 9999
		if width > 0 && width <= max && height == 0 {
			suffix += fmt.Sprintf("%dx", width)
		}
		if width == 0 && height > 0 && height <= max {
			suffix += fmt.Sprintf("x%d", height)
		}
		if width > 0 && height > 0 && width <= max && height <= max {
			suffix += fmt.Sprintf("%dx%d!", width, height)
		}
	}

	return suffix
}
