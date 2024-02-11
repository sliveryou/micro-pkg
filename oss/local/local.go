package local

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/sliveryou/go-tool/v2/filex"
)

const (
	// CloudLocal 云服务商：本地
	CloudLocal = "local"
)

// OptionFunc 可选配置
type OptionFunc func(l *LSS)

// WithSecure 使用安全配置
func WithSecure(secure ...bool) OptionFunc {
	return func(l *LSS) {
		l.secure = true
		if len(secure) > 0 {
			l.secure = secure[0]
		}
	}
}

// LSS 本地 LSS 结构详情
type LSS struct {
	secure     bool   // 是否使用安全配置
	endpoint   string // 端节点
	bucketName string // 存储桶名称
}

// NewLSS 创建一个本地 LSS 对象
func NewLSS(endpoint, bucketName string, opts ...OptionFunc) (*LSS, error) {
	if bucketName == "" {
		return nil, errors.New("local: illegal lss config")
	}

	l := &LSS{endpoint: endpoint, bucketName: bucketName}
	for _, opt := range opts {
		opt(l)
	}

	return l, nil
}

// Cloud 获取云服务商名称
func (l *LSS) Cloud() string {
	return CloudLocal
}

// GetURL 获取对象在本地 LSS 上的完整访问 URL
func (l *LSS) GetURL(key string) string {
	if strings.HasPrefix(key, "http") {
		return key
	}

	if l.endpoint != "" {
		scheme := "http"
		if l.secure {
			scheme = "https"
		}
		return fmt.Sprintf("%s://%s/%s", scheme, l.endpoint, key)
	}

	return key
}

// GetObject 获取对象在本地 LSS 的存储数据
func (l *LSS) GetObject(key string) (io.ReadCloser, error) {
	destPath := filepath.Join(l.bucketName, key)
	obj, err := os.Open(destPath)
	if err != nil {
		return nil, errors.WithMessage(err, "local: lss get object err")
	}

	return obj, nil
}

// PutObject 上传对象至本地 LSS
func (l *LSS) PutObject(key string, reader io.Reader) (string, error) {
	destPath := filepath.Join(l.bucketName, key)
	if err := l.mkdir(destPath); err != nil {
		return "", err
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return "", errors.WithMessagef(err, "local: lss create dest path = %s err", destPath)
	}
	defer dest.Close()

	_, err = io.Copy(dest, reader)
	if err != nil {
		return "", errors.WithMessage(err, "local: lss copy reader to dest err")
	}

	return l.GetURL(key), nil
}

// DeleteObjects 批量删除本地 LSS 上的对象
func (l *LSS) DeleteObjects(keys ...string) error {
	for _, key := range keys {
		destPath := filepath.Join(l.bucketName, key)
		_ = os.Remove(destPath)
	}

	return nil
}

// UploadFile 上传文件至本地 LSS，filePath：文件路径，partSize：分块大小（字节），routines：并发数
func (l *LSS) UploadFile(key, filePath string, partSize int64, routines int) (string, error) {
	destPath := filepath.Join(l.bucketName, key)
	if err := l.mkdir(destPath); err != nil {
		return "", err
	}

	if filePath != destPath {
		src, err := os.Open(filePath)
		if err != nil {
			return "", errors.WithMessagef(err, "local: lss open file path = %v err", filePath)
		}
		defer src.Close()

		dest, err := os.Create(destPath)
		if err != nil {
			return "", errors.WithMessagef(err, "local: lss create dest path = %v err", destPath)
		}
		defer dest.Close()

		_, err = io.Copy(dest, src)
		if err != nil {
			return "", errors.WithMessage(err, "local: lss copy src to dest err")
		}
	}

	return l.GetURL(key), nil
}

// AuthorizedUpload 授权上传至本地 LSS，expires：过期时间（秒）
func (l *LSS) AuthorizedUpload(key string, expires int) (string, error) {
	return l.GetURL(key), nil
}

// GetThumbnailSuffix 获取缩略图后缀
func (l *LSS) GetThumbnailSuffix(width, height int, size int64) string {
	return ""
}

// mkdir 创建目标目录
func (l *LSS) mkdir(destPath string) error {
	destDir := filepath.Dir(destPath)
	if !filex.IsExist(destDir) {
		return errors.WithMessagef(filex.Mkdir(destDir),
			"local lss mkdir dest dir = %s err", destDir)
	}

	return nil
}
