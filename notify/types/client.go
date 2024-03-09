package types

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/sliveryou/micro-pkg/xhttp"
)

// NewSmsClientPicker 新建短信客户端选取器
func NewSmsClientPicker() SmsClientPicker {
	return &smsClientPicker{
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		pickMap: make(map[string]SmsClient),
	}
}

// smsClientPicker 短信客户端选取器
type smsClientPicker struct {
	mu      sync.RWMutex         // 读写锁
	r       *rand.Rand           // 随机源
	pickMap map[string]SmsClient // 短信客户端选取 map
}

// Pick 选取一个短信客户端
func (p *smsClientPicker) Pick() (sc SmsClient, key string, isExist bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	rn := p.r.Intn(len(p.pickMap))
	for k := range p.pickMap {
		if rn == 0 {
			key = k
		}
		rn--
	}

	sc, isExist = p.pickMap[key]
	return
}

// Get 获取一个短信客户端
func (p *smsClientPicker) Get(key string) (sc SmsClient, isExist bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sc, isExist = p.pickMap[key]
	return
}

// Add 添加一个短信客户端
func (p *smsClientPicker) Add(key string, value SmsClient) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pickMap[key] = value
}

// Remove 移除一个短信客户端
func (p *smsClientPicker) Remove(keys ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, key := range keys {
		delete(p.pickMap, key)
	}
}

// NewEmailClientPicker 新建邮件客户端选取器
func NewEmailClientPicker() EmailClientPicker {
	return &emailClientPicker{
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		pickMap: make(map[string]EmailClient),
	}
}

// emailClientPicker 邮件客户端选取器
type emailClientPicker struct {
	mu      sync.RWMutex           // 读写锁
	r       *rand.Rand             // 随机源
	pickMap map[string]EmailClient // 邮件客户端选取map
}

// Pick 选取一个邮件客户端
func (p *emailClientPicker) Pick() (ec EmailClient, key string, isExist bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	rn := p.r.Intn(len(p.pickMap))
	for k := range p.pickMap {
		if rn == 0 {
			key = k
		}
		rn--
	}

	ec, isExist = p.pickMap[key]
	return
}

// Get 获取一个邮件客户端
func (p *emailClientPicker) Get(key string) (ec EmailClient, isExist bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ec, isExist = p.pickMap[key]
	return
}

// Add 添加一个邮件客户端
func (p *emailClientPicker) Add(key string, value EmailClient) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pickMap[key] = value
}

// Remove 移除一个邮件客户端
func (p *emailClientPicker) Remove(keys ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, key := range keys {
		delete(p.pickMap, key)
	}
}

// Option 可选配置
type Option func(bc *BaseClient)

// WithHTTPClient 使用 HTTP 客户端
func WithHTTPClient(hc *http.Client) Option {
	return func(bc *BaseClient) {
		if hc != nil {
			bc.HTTPClient = hc
		}
	}
}

// WithSmsTmplMap 使用短信对应模板映射
func WithSmsTmplMap(m map[string]string) Option {
	return func(bc *BaseClient) {
		if m != nil {
			bc.smsTmplMap = m
		}
	}
}

// WithEmailTmplMap 使用邮件对应模板映射
func WithEmailTmplMap(m map[string]string) Option {
	return func(bc *BaseClient) {
		if m != nil {
			bc.emailTmplMap = m
		}
	}
}

// NewBaseClient 新建基础客户端
func NewBaseClient(opts ...Option) *BaseClient {
	bc := &BaseClient{}

	for _, opt := range opts {
		opt(bc)
	}

	if bc.HTTPClient == nil {
		bc.HTTPClient = xhttp.NewHTTPClient()
	}

	return bc
}

// BaseClient 基础客户端
type BaseClient struct {
	HTTPClient   *http.Client      // HTTP 客户端
	smsTmplMap   map[string]string // 短信对应模板映射
	emailTmplMap map[string]string // 邮件对应模板映射
}

// ParseSmsTmpl 解析短信对应模板
func (bc *BaseClient) ParseSmsTmpl(templateID string) string {
	parsed := templateID
	if bc.smsTmplMap != nil {
		if t, ok := bc.smsTmplMap[templateID]; ok && t != "" {
			parsed = t
		}
	}

	return parsed
}

// ParseEmailTmpl 解析邮件对应模板
func (bc *BaseClient) ParseEmailTmpl(templateID string) string {
	parsed := templateID
	if bc.emailTmplMap != nil {
		if t, ok := bc.emailTmplMap[templateID]; ok && t != "" {
			parsed = t
		}
	}

	return parsed
}

// MockClient 模拟短信、邮件客户端
type MockClient struct{}

// Platform 服务平台
func (c *MockClient) Platform() string {
	return "mock"
}

// SendSms 发送短信
func (c *MockClient) SendSms(receiver, templateID string, params ...Param) error {
	var b strings.Builder
	b.WriteByte('[')
	for i, p := range params {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(strings.TrimPrefix(fmt.Sprintf("%+v", p), "&"))
	}
	b.WriteByte(']')

	logx.Infof("notify: mock client send sms, receiver: %s, template id: %s, params: %s",
		receiver, templateID, b.String())

	return nil
}

// SendEmail 发送邮件
func (c *MockClient) SendEmail(receiver, templateID string, params ...Param) error {
	var b strings.Builder
	b.WriteByte('[')
	for i, p := range params {
		if i != 0 {
			b.WriteString(", ")
		}
		b.WriteString(strings.TrimPrefix(fmt.Sprintf("%+v", p), "&"))
	}
	b.WriteByte(']')

	logx.Infof("notify: mock client send email, receiver: %s, template id: %s, params: %s",
		receiver, templateID, b.String())

	return nil
}
