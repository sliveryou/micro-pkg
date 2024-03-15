package xhttp

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
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
	*http.Client
}

// NewClient 新建 HTTP 拓展客户端
func NewClient(config ...Config) *Client {
	return &Client{Client: NewHTTPClient(config...)}
}

// NewClientWithHTTPClient 使用 HTTP 客户端新建 HTTP 拓展客户端
func NewClientWithHTTPClient(client *http.Client) *Client {
	if client == nil {
		panic(errors.New("nil client is invalid"))
	}

	return &Client{Client: client}
}

// GetRequest 获取 HTTP 请求
func (c *Client) GetRequest(ctx context.Context, method, url string, header map[string]string, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, data)
	if err != nil {
		return nil, errors.WithMessagef(err, "new http request err, method: %s, url: %s, header: %v",
			method, url, header)
	}

	for k, v := range header {
		if k == HeaderHost && v != "" {
			req.Host = v
		} else {
			req.Header.Set(k, v)
		}
	}

	return req, nil
}

// GetResponse 获取 HTTP 响应及其响应体内容
func (c *Client) GetResponse(req *http.Request) (*http.Response, []byte, error) {
	response, err := c.Do(req)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "http client do request err")
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "read all response body err")
	}

	return response, body, nil
}

// CallWithRequest 利用 HTTP 请求进行 HTTP 调用，并将 json 响应内容反序列化至 resp 中
func (c *Client) CallWithRequest(req *http.Request, resp any) (*http.Response, error) {
	response, body, err := c.GetResponse(req)
	if err != nil {
		return nil, errors.WithMessage(err, "get response err")
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, resp)
		if err != nil {
			return nil, errors.WithMessage(err, "json unmarshal response body err")
		}
	}

	return response, nil
}

// Call HTTP 调用，并将 json 响应内容反序列化至 resp 中
func (c *Client) Call(ctx context.Context, method, url string, header map[string]string, data io.Reader, resp any) (*http.Response, error) {
	req, err := c.GetRequest(ctx, method, url, header, data)
	if err != nil {
		return nil, errors.WithMessage(err, "get request err")
	}

	return c.CallWithRequest(req, resp)
}
