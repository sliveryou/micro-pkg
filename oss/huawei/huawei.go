package huawei

import (
	"fmt"
	"io"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/xhttp"
)

const (
	// CloudHuawei 云服务商：华为云
	CloudHuawei = "huawei"
)

// OptionFunc 可选配置
type OptionFunc func(o *OBS)

// WithNotSetACL 不设置权限规则
func WithNotSetACL(notSetACL ...bool) OptionFunc {
	return func(o *OBS) {
		o.notSetACL = true
		if len(notSetACL) > 0 {
			o.notSetACL = notSetACL[0]
		}
	}
}

// OBS 华为云 OBS 结构详情
type OBS struct {
	notSetACL  bool
	endpoint   string
	bucketName string
	client     *obs.ObsClient
}

// NewOBS 创建一个华为云 OBS 对象
func NewOBS(endpoint, accessKeyID, accessKeySecret, bucketName string, opts ...OptionFunc) (*OBS, error) {
	client, err := obs.New(accessKeyID, accessKeySecret, endpoint)
	if err != nil {
		return nil, errors.WithMessage(err, "huawei: new obs err")
	}

	o := &OBS{endpoint: endpoint, bucketName: bucketName, client: client}
	for _, opt := range opts {
		opt(o)
	}
	if !o.notSetACL {
		// 设置存储桶公共读权限规则
		_, err = client.SetBucketAcl(&obs.SetBucketAclInput{Bucket: bucketName, ACL: obs.AclPublicRead})
		if err != nil {
			return nil, errors.WithMessage(err, "huawei: obs set bucket acl err")
		}
	}

	return o, nil
}

// Cloud 获取云服务商名称
func (o *OBS) Cloud() string {
	return CloudHuawei
}

// GetURL 获取对象在华为云 OBS 上的完整访问 URL
func (o *OBS) GetURL(key string) string {
	if strings.HasPrefix(key, "http") {
		return key
	}

	return fmt.Sprintf("https://%s.%s/%s", o.bucketName, o.endpoint, key)
}

// GetObject 获取对象在华为云 OBS 的存储数据
func (o *OBS) GetObject(key string) (io.ReadCloser, error) {
	input := &obs.GetObjectInput{}
	input.Bucket = o.bucketName
	input.Key = key

	obj, err := o.client.GetObject(input)
	if err != nil {
		return nil, errors.WithMessage(err, "huawei: obs get object err")
	}

	return obj.Body, nil
}

// PutObject 上传对象至华为云 OBS
func (o *OBS) PutObject(key string, reader io.Reader) (string, error) {
	input := &obs.PutObjectInput{}
	input.Bucket = o.bucketName
	input.Key = key
	input.Body = reader
	input.ContentType = xhttp.TypeByExtension(key)
	input.ContentLength, _ = xhttp.GetReaderLen(reader)

	_, err := o.client.PutObject(input)
	if err != nil {
		return "", errors.WithMessage(err, "huawei: obs put object err")
	}

	return o.GetURL(key), nil
}

// DeleteObjects 批量删除华为云 OBS 上的对象
func (o *OBS) DeleteObjects(keys ...string) error {
	deletes := make([]obs.ObjectToDelete, 0, len(keys))
	for _, object := range keys {
		deletes = append(deletes, obs.ObjectToDelete{Key: object})
	}

	input := &obs.DeleteObjectsInput{}
	input.Bucket = o.bucketName
	input.Objects = deletes

	_, err := o.client.DeleteObjects(input)
	if err != nil {
		return errors.WithMessage(err, "huawei: obs delete objects err")
	}

	return nil
}

// UploadFile 上传文件至华为云 OBS，filePath：文件路径，partSize：分块大小（字节），routines：并发数
func (o *OBS) UploadFile(key, filePath string, partSize int64, routines int) (string, error) {
	input := &obs.UploadFileInput{}
	input.Bucket = o.bucketName
	input.Key = key
	input.UploadFile = filePath
	input.EnableCheckpoint = true // 开启断点续传模式
	input.PartSize = partSize     // 指定分段大小
	input.TaskNum = routines      // 指定分段上传时的最大并发数
	input.ContentType = xhttp.TypeByExtension(key)

	_, err := o.client.UploadFile(input)
	if err != nil {
		return "", errors.WithMessage(err, "huawei: obs upload file err")
	}

	return o.GetURL(key), nil
}

// AuthorizedUpload 授权上传至华为云 OBS，expires：过期时间（秒）
func (o *OBS) AuthorizedUpload(key string, expires int) (string, error) {
	input := &obs.CreateSignedUrlInput{}
	input.Bucket = o.bucketName
	input.Key = key
	input.Expires = expires
	input.Method = obs.HttpMethodPut

	output, err := o.client.CreateSignedUrl(input)
	if err != nil {
		return "", errors.WithMessage(err, "huawei: obs authorized upload err")
	}

	return output.SignedUrl, nil
}

// GetThumbnailSuffix 获取缩略图后缀
func (o *OBS) GetThumbnailSuffix(width, height int, size int64) string {
	// 参考文档 https://support.huaweicloud.com/fg-obs/obs_01_0430.html
	var suffix string
	if height > 0 || width > 0 {
		suffix = "?x-image-process=image/resize,limit_0"
		max := 4096
		if width > 0 && width <= max {
			suffix += fmt.Sprintf(",w_%d", width)
		}

		if height > 0 && height <= max {
			suffix += fmt.Sprintf(",h_%d", height)
		}
	}

	return suffix
}
