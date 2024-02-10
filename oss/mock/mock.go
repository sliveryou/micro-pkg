package mock

import (
	"io"
)

const (
	// CloudMock 云服务商：模拟
	CloudMock = "mock"
)

// MSS 模拟 MSS 结构详情
type MSS struct{}

// NewMSS 创建一个模拟 MSS 对象
func NewMSS() *MSS {
	return &MSS{}
}

// Cloud 获取云服务商名称
func (m *MSS) Cloud() string {
	return CloudMock
}

// GetURL 获取对象在模拟 MSS 上的完整访问 URL
func (m *MSS) GetURL(key string) string {
	return key
}

// GetObject 获取对象在模拟 MSS 的存储数据
func (m *MSS) GetObject(key string) (io.ReadCloser, error) {
	return &mockReadCloser{}, nil
}

// PutObject 上传对象至模拟 MSS
func (m *MSS) PutObject(key string, reader io.Reader) (string, error) {
	return m.GetURL(key), nil
}

// DeleteObjects 批量删除模拟 MSS 上的对象
func (m *MSS) DeleteObjects(keys ...string) error {
	return nil
}

// UploadFile 上传文件至模拟 MSS，filePath：文件路径，partSize：分块大小（字节），routines：并发数
func (m *MSS) UploadFile(key, filePath string, partSize int64, routines int) (string, error) {
	return m.GetURL(key), nil
}

// AuthorizedUpload 授权上传至模拟 MSS，expires：过期时间（秒）
func (m *MSS) AuthorizedUpload(key string, expires int) (string, error) {
	return m.GetURL(key), nil
}

// GetThumbnailSuffix 获取缩略图后缀
func (m *MSS) GetThumbnailSuffix(width, height int, size int64) string {
	return ""
}

// mockReadCloser 模拟 io.ReadCloser
type mockReadCloser struct{}

// Read 实现Read方法
func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

// Close 实现Close方法
func (m *mockReadCloser) Close() error {
	return nil
}
