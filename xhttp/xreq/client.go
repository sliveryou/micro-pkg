package xreq

import (
	"net"
	"net/http"
	"time"
)

var (
	// DefaultHTTPClient 默认 HTTP 客户端
	DefaultHTTPClient = NewHTTPClient()
	// DefaultClient 默认 HTTP 拓展客户端
	DefaultClient = NewClient(DefaultHTTPClient)
)

// Config HTTP 客户端相关配置
type Config struct {
	HTTPTimeout           time.Duration // HTTP 请求超时时间
	DialTimeout           time.Duration // 拨号超时时间
	DialKeepAlive         time.Duration // 拨号保持连接时间
	MaxIdleConns          int           // 最大空闲连接数
	MaxIdleConnsPerHost   int           // 每个主机最大空闲连接数
	IdleConnTimeout       time.Duration // 空闲连接超时时间
	TLSHandshakeTimeout   time.Duration // TLS 握手超时时间
	ExpectContinueTimeout time.Duration // 期望继续超时时间
}

// DefaultConfig 获取默认 HTTP 客户端相关配置
func DefaultConfig() Config {
	return Config{
		HTTPTimeout:           20 * time.Second,
		DialTimeout:           15 * time.Second,
		DialKeepAlive:         30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 2 * time.Second,
	}
}

// NewHTTPClient 新建 HTTP 客户端（不传递配置时，将使用默认配置 DefaultConfig）
func NewHTTPClient(config ...Config) *http.Client {
	c := DefaultConfig()
	if len(config) > 0 {
		c = config[0]
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   c.DialTimeout,
			KeepAlive: c.DialKeepAlive,
		}).DialContext,
		MaxIdleConns:          c.MaxIdleConns,
		IdleConnTimeout:       c.IdleConnTimeout,
		MaxIdleConnsPerHost:   c.MaxIdleConnsPerHost,
		TLSHandshakeTimeout:   c.TLSHandshakeTimeout,
		ExpectContinueTimeout: c.ExpectContinueTimeout,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   c.HTTPTimeout,
	}
}

// Client HTTP 拓展客户端结构详情
type Client struct {
	OptionCollection
	doer *http.Client
}

// NewClient 新建拓展 HTTP 客户端
func NewClient(doer *http.Client, options ...Option) *Client {
	if doer == nil {
		doer = DefaultHTTPClient
	}

	return &Client{doer: doer, OptionCollection: options}
}

// Do 发起 HTTP 请求
func (c *Client) Do(method string, options ...Option) (*http.Response, error) {
	request, err := New(method, "", append(c.OptionCollection, options...)...)
	if err != nil {
		return nil, err
	}

	return c.doer.Do(request)
}

// DoWithRequest 使用 *http.Request 发起 HTTP 请求
func (c *Client) DoWithRequest(request *http.Request, options ...Option) (*http.Response, error) {
	var err error
	request, err = Apply(request, options...)
	if err != nil {
		return nil, err
	}

	return c.doer.Do(request)
}

// Get 发起 HTTP GET 请求
func (c *Client) Get(options ...Option) (*http.Response, error) {
	return c.Do(http.MethodGet, options...)
}

// Post 发起 HTTP POST 请求
func (c *Client) Post(options ...Option) (*http.Response, error) {
	return c.Do(http.MethodPost, options...)
}

// Put 发起 HTTP PUT 请求
func (c *Client) Put(options ...Option) (*http.Response, error) {
	return c.Do(http.MethodPut, options...)
}

// Patch 发起 HTTP PATCH 请求
func (c *Client) Patch(options ...Option) (*http.Response, error) {
	return c.Do(http.MethodPatch, options...)
}

// Delete 发起 HTTP DELETE 请求
func (c *Client) Delete(options ...Option) (*http.Response, error) {
	return c.Do(http.MethodDelete, options...)
}

// Head 发起 HTTP HEAD 请求
func (c *Client) Head(options ...Option) (*http.Response, error) {
	return c.Do(http.MethodHead, options...)
}

// Options 发起 HTTP OPTIONS 请求
func (c *Client) Options(options ...Option) (*http.Response, error) {
	return c.Do(http.MethodOptions, options...)
}
