package xhttp

import (
	"bytes"
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
	MaxConnsPerHost       int           // 每个主机最大连接数
	IdleConnTimeout       time.Duration // 空闲连接超时时间
	ResponseHeaderTimeout time.Duration // 读取响应头超时时间
	ExpectContinueTimeout time.Duration // 期望继续超时时间
	TLSHandshakeTimeout   time.Duration // TLS 握手超时时间
	ForceAttemptHTTP2     bool          // 允许尝试启用 HTTP/2
}

// GetDefaultConfig 获取默认 HTTP 客户端相关配置
func GetDefaultConfig() *Config {
	return &Config{
		HTTPTimeout:           20 * time.Second,
		DialTimeout:           15 * time.Second,
		DialKeepAlive:         30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		MaxConnsPerHost:       100,
		IdleConnTimeout:       60 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ForceAttemptHTTP2:     true,
	}
}

// NewHTTPClient 新建 HTTP 客户端
func NewHTTPClient(c *Config) *http.Client {
	if c == nil {
		c = GetDefaultConfig()
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   c.DialTimeout,
			KeepAlive: c.DialKeepAlive,
		}).DialContext,
		MaxIdleConns:          c.MaxIdleConns,
		MaxIdleConnsPerHost:   c.MaxIdleConnsPerHost,
		MaxConnsPerHost:       c.MaxConnsPerHost,
		IdleConnTimeout:       c.IdleConnTimeout,
		ResponseHeaderTimeout: c.ResponseHeaderTimeout,
		ExpectContinueTimeout: c.ExpectContinueTimeout,
		TLSHandshakeTimeout:   c.TLSHandshakeTimeout,
		ForceAttemptHTTP2:     c.ForceAttemptHTTP2,
	}

	client := &http.Client{
		Timeout:   c.HTTPTimeout,
		Transport: tr,
	}

	return client
}

// NewDefaultHTTPClient 新建默认 HTTP 客户端
func NewDefaultHTTPClient() *http.Client {
	return NewHTTPClient(nil)
}

// Client HTTP 拓展客户端结构详情
type Client struct {
	*http.Client
}

// NewClient 新建 HTTP 拓展客户端
func NewClient(c *Config) *Client {
	return &Client{Client: NewHTTPClient(c)}
}

// NewDefaultClient 新建默认 HTTP 拓展客户端
func NewDefaultClient() *Client {
	return &Client{Client: NewDefaultHTTPClient()}
}

// NewClientWithHTTPClient 使用 HTTP 客户端新建 HTTP 拓展客户端
func NewClientWithHTTPClient(client *http.Client) *Client {
	return &Client{Client: client}
}

// GetRequest 获取 HTTP 请求
func (c *Client) GetRequest(ctx context.Context, method, url string, header map[string]string, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, data)
	if err != nil {
		return nil, errors.WithMessagef(err, "new http request err, method = %v, url = %v, header = %v",
			method, url, header)
	}

	for k, v := range header {
		req.Header.Add(k, v)
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

// CallWithRequest 利用 HTTP 请求进行 HTTP 调用
func (c *Client) CallWithRequest(req *http.Request, resp interface{}) (*http.Response, error) {
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

// Call HTTP 调用
func (c *Client) Call(ctx context.Context, method, url string, header map[string]string, data io.Reader, resp interface{}) error {
	req, err := c.GetRequest(ctx, method, url, header, data)
	if err != nil {
		return errors.WithMessage(err, "get request err")
	}

	_, err = c.CallWithRequest(req, resp)
	if err != nil {
		return errors.WithMessage(err, "call with request err")
	}

	return nil
}

// ChainReq 区块链 HTTP 调用请求
type ChainReq struct {
	ID     int         `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// ChainResp 区块链 HTTP 调用响应
type ChainResp struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

// CallChain 区块链HTTP调用
func (c *Client) CallChain(ctx context.Context, method, url string, header map[string]string, params, resp interface{}) error {
	cReq := &ChainReq{Method: method, Params: params}
	data, err := json.Marshal(cReq)
	if err != nil {
		return errors.WithMessage(err, "json marshal chain request err")
	}

	req, err := c.GetRequest(ctx, http.MethodPost, url, header, bytes.NewBuffer(data))
	if err != nil {
		return errors.WithMessage(err, "get request err")
	}

	cResp := ChainResp{}
	_, err = c.CallWithRequest(req, &cResp)
	if err != nil {
		return errors.WithMessage(err, "call with request err")
	}

	if cResp.Error != "" {
		return errors.Errorf("chain return err: %v", cResp.Error)
	}

	err = json.Unmarshal(cResp.Result, resp)
	if err != nil {
		return errors.WithMessage(err, "json unmarshal chain result err")
	}

	return nil
}
