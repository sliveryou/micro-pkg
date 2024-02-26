package minio

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/xhttp"
)

const (
	// CloudMinIO 云服务商：MinIO
	CloudMinIO = "minio"

	// readPolicy 存储桶公共读权限规则
	readPolicy = `{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Principal": {
			"AWS": ["*"]
		},
		"Action": ["s3:GetBucketLocation"],
		"Resource": ["arn:aws:s3:::${BUCKET_NAME}"]
	}, {
		"Effect": "Allow",
		"Principal": {
			"AWS": ["*"]
		},
		"Action": ["s3:GetObject"],
		"Resource": ["arn:aws:s3:::${BUCKET_NAME}/*"]
	}]
}`
)

// Option 可选配置
type Option func(m *MinIO)

// WithSecure 使用安全配置
func WithSecure(secure ...bool) Option {
	return func(m *MinIO) {
		m.secure = true
		if len(secure) > 0 {
			m.secure = secure[0]
		}
	}
}

// WithNotSetACL 不设置权限规则
func WithNotSetACL(notSetACL ...bool) Option {
	return func(m *MinIO) {
		m.notSetACL = true
		if len(notSetACL) > 0 {
			m.notSetACL = notSetACL[0]
		}
	}
}

// MinIO 结构详情
type MinIO struct {
	secure     bool
	notSetACL  bool
	endpoint   string
	bucketName string
	client     *minio.Client
}

// NewMinIO 创建一个 MinIO 对象
func NewMinIO(endpoint, accessKeyID, accessKeySecret, bucketName string, opts ...Option) (*MinIO, error) {
	m := &MinIO{endpoint: endpoint, bucketName: bucketName}
	for _, opt := range opts {
		opt(m)
	}

	client, err := minio.New(m.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, accessKeySecret, ""),
		Secure: m.secure,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "minio: new minio err")
	}

	if !m.notSetACL {
		// 设置存储桶公共读权限规则
		err = client.SetBucketPolicy(context.Background(), m.bucketName,
			strings.ReplaceAll(readPolicy, "${BUCKET_NAME}", m.bucketName))
		if err != nil {
			return nil, errors.WithMessage(err, "minio: set bucket policy err")
		}
	}
	m.client = client

	return m, nil
}

// Cloud 获取云服务商名称
func (m *MinIO) Cloud() string {
	return CloudMinIO
}

// GetURL 获取对象在 MinIO 上的完整访问 URL
func (m *MinIO) GetURL(key string) string {
	if strings.HasPrefix(key, "http") {
		return key
	}

	scheme := "http"
	if m.secure {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, m.endpoint, m.bucketName, key)
}

// GetObject 获取对象在 MinIO 的存储数据
func (m *MinIO) GetObject(key string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(context.Background(), m.bucketName, key,
		minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.WithMessage(err, "minio: get object err")
	}

	return obj, nil
}

// PutObject 上传对象至 MinIO
func (m *MinIO) PutObject(key string, reader io.Reader) (string, error) {
	contentLen, _ := xhttp.GetReaderLen(reader)
	if contentLen == 0 {
		contentLen = -1
	}

	_, err := m.client.PutObject(context.Background(), m.bucketName, key, reader, contentLen,
		minio.PutObjectOptions{
			ContentType: xhttp.TypeByExtension(key),
		})
	if err != nil {
		return "", errors.WithMessage(err, "minio: put object err")
	}

	return m.GetURL(key), nil
}

// DeleteObjects 批量删除 MinIO 上的对象
func (m *MinIO) DeleteObjects(keys ...string) error {
	objectCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectCh)
		for _, key := range keys {
			objectCh <- minio.ObjectInfo{Key: key}
		}
	}()

	var err error
	for errCh := range m.client.RemoveObjects(context.Background(), m.bucketName, objectCh,
		minio.RemoveObjectsOptions{}) {
		if errCh.Err != nil {
			err = errors.WithMessagef(err, "remove object = %v err", errCh.ObjectName)
		}
	}

	return errors.WithMessage(err, "minio: delete objects err")
}

// UploadFile 上传文件至 MinIO，filePath：文件路径，partSize：分块大小（字节），routines：并发数
func (m *MinIO) UploadFile(key, filePath string, partSize int64, routines int) (string, error) {
	_, err := m.client.FPutObject(context.Background(), m.bucketName, key, filePath,
		minio.PutObjectOptions{
			ContentType: xhttp.TypeByExtension(filePath),
			PartSize:    uint64(partSize),
			NumThreads:  uint(routines),
		})
	if err != nil {
		return "", errors.WithMessage(err, "minio: upload file err")
	}

	return m.GetURL(key), nil
}

// AuthorizedUpload 授权上传至 MinIO，expires：过期时间（秒）
func (m *MinIO) AuthorizedUpload(key string, expires int) (string, error) {
	signedURL, err := m.client.PresignedPutObject(context.Background(), m.bucketName, key,
		time.Duration(expires)*time.Second)
	if err != nil {
		return "", errors.WithMessage(err, "minio: authorized upload err")
	}

	return signedURL.String(), nil
}

// GetThumbnailSuffix 获取缩略图后缀
func (m *MinIO) GetThumbnailSuffix(width, height int, size int64) string {
	return ""
}
