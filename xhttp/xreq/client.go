package xreq

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/xhttp"
)

var (
	// DefaultConfig 获取默认 HTTP 客户端配置
	DefaultConfig = xhttp.DefaultConfig
	// NewHTTPClient 新建 HTTP 客户端（不传递配置时，将使用默认配置 DefaultConfig）
	NewHTTPClient = xhttp.NewHTTPClient

	// DefaultHTTPClient 默认 HTTP 客户端
	DefaultHTTPClient = NewHTTPClient()
	// DefaultClient 默认 HTTP 拓展客户端
	DefaultClient = NewClient()
)

// Config HTTP 客户端配置
type Config = xhttp.Config

// Client HTTP 拓展客户端
type Client struct {
	BeforeOptions OptionCollection
	AfterOptions  OptionCollection
	JSONUnmarshal Unmarshaler
	XMLUnmarshal  Unmarshaler
	httpClient    *http.Client
}

// NewClient 新建 HTTP 拓展客户端
func NewClient(beforeOptions ...Option) *Client {
	return createClient(nil, beforeOptions...)
}

// NewClientWithConfig 使用配置新建 HTTP 拓展客户端
func NewClientWithConfig(config Config, beforeOptions ...Option) *Client {
	return createClient(NewHTTPClient(config), beforeOptions...)
}

// NewClientWithHTTPClient 使用 HTTP 客户端新建 HTTP 拓展客户端
func NewClientWithHTTPClient(hc *http.Client, beforeOptions ...Option) *Client {
	return createClient(hc, beforeOptions...)
}

// Do 发起 HTTP 请求
func (c *Client) Do(method string, options ...Option) (*Response, error) {
	request, err := New(method, "", c.concatOptions(options...)...)
	if err != nil {
		return nil, err
	}

	return c.roundTrip(request)
}

// Call 发起 HTTP 请求，并根据响应头部 "Content-Type"
// 的值将响应体内容使用特定 Unmarshaler 函数反序列化至 result 中
func (c *Client) Call(method string, result any, options ...Option) (*Response, error) {
	response, err := c.Do(method, options...)
	if err != nil {
		return nil, err
	}

	err = response.Unmarshal(result)

	return response, err
}

// DoWithRequest 使用 *http.Request 发起 HTTP 请求
func (c *Client) DoWithRequest(request *http.Request, options ...Option) (*Response, error) {
	var err error
	request, err = Apply(request, c.concatOptions(options...)...)
	if err != nil {
		return nil, err
	}

	return c.roundTrip(request)
}

// CallWithRequest 使用 *http.Request 发起 HTTP 请求，并根据响应头部
// "Content-Type" 的值将响应体内容使用特定 Unmarshaler 函数反序列化至 result 中
func (c *Client) CallWithRequest(request *http.Request, result any, options ...Option) (*Response, error) {
	response, err := c.DoWithRequest(request, options...)
	if err != nil {
		return nil, err
	}

	err = response.Unmarshal(result)

	return response, err
}

// Get 发起 HTTP GET 请求
func (c *Client) Get(options ...Option) (*Response, error) {
	return c.Do(http.MethodGet, options...)
}

// Post 发起 HTTP POST 请求
func (c *Client) Post(options ...Option) (*Response, error) {
	return c.Do(http.MethodPost, options...)
}

// Put 发起 HTTP PUT 请求
func (c *Client) Put(options ...Option) (*Response, error) {
	return c.Do(http.MethodPut, options...)
}

// Patch 发起 HTTP PATCH 请求
func (c *Client) Patch(options ...Option) (*Response, error) {
	return c.Do(http.MethodPatch, options...)
}

// Delete 发起 HTTP DELETE 请求
func (c *Client) Delete(options ...Option) (*Response, error) {
	return c.Do(http.MethodDelete, options...)
}

// Head 发起 HTTP HEAD 请求
func (c *Client) Head(options ...Option) (*Response, error) {
	return c.Do(http.MethodHead, options...)
}

// Options 发起 HTTP OPTIONS 请求
func (c *Client) Options(options ...Option) (*Response, error) {
	return c.Do(http.MethodOptions, options...)
}

// GetHTTPClient 获取 HTTP 客户端
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

// SetHTTPClient 设置 HTTP 客户端
func (c *Client) SetHTTPClient(hc *http.Client) *Client {
	if hc != nil {
		c.httpClient = hc
	}

	return c
}

// SetProxy 设置请求代理
func (c *Client) SetProxy(proxyURL string) *Client {
	transport, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		return c
	}

	pURL, err := url.Parse(proxyURL)
	if err != nil {
		return c
	}

	transport.Proxy = http.ProxyURL(pURL)

	return c
}

// SetTransport 设置传输器
func (c *Client) SetTransport(transport http.RoundTripper) *Client {
	if transport != nil {
		c.httpClient.Transport = transport
	}

	return c
}

// SetTLSClientConfig 设置 TLS 配置
func (c *Client) SetTLSClientConfig(config *tls.Config) *Client {
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = config
	}

	return c
}

// SetTimeout 设置 HTTP 请求超时时间
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout

	return c
}

// SetAfterOptions 设置后可选参数
func (c *Client) SetAfterOptions(afterOptions ...Option) *Client {
	c.AfterOptions = afterOptions

	return c
}

// SetJSONUnmarshaler 设置 json 反序列化器
func (c *Client) SetJSONUnmarshaler(unmarshaler Unmarshaler) *Client {
	c.JSONUnmarshal = unmarshaler

	return c
}

// SetXMLUnmarshaler 设置 xml 反序列化器
func (c *Client) SetXMLUnmarshaler(unmarshaler Unmarshaler) *Client {
	c.XMLUnmarshal = unmarshaler

	return c
}

// concatOptions 连接可选参数
func (c *Client) concatOptions(options ...Option) []Option {
	return c.BeforeOptions.With(options...).With(c.AfterOptions...)
}

// roundTrip 请求响应往返
func (c *Client) roundTrip(request *http.Request) (*Response, error) {
	resp, err := c.httpClient.Do(request)
	response := &Response{RawResponse: resp, client: c}
	if err != nil {
		response.setReceivedAt()
		return response, errors.WithMessagef(err, "do http request err")
	}

	defer resp.Body.Close()

	body, contentEncoding := resp.Body, resp.Header.Get(xhttp.HeaderContentEncoding)
	if strings.EqualFold(contentEncoding, "gzip") && resp.ContentLength != 0 {
		if _, ok := body.(*gzip.Reader); !ok {
			body, err = gzip.NewReader(body)
			if err != nil {
				response.setReceivedAt()
				return response, errors.WithMessagef(err, "new gzip reader err")
			}

			defer body.Close()
		}
	}

	if response.body, err = io.ReadAll(body); err != nil {
		response.setReceivedAt()
		return response, errors.WithMessage(err, "real all body err")
	}

	resp.Body = rc(bytes.NewReader(response.body))
	response.size = int64(len(response.body))
	response.setReceivedAt()

	return response, nil
}

// createClient 创建 HTTP 拓展客户端
func createClient(hc *http.Client, beforeOptions ...Option) *Client {
	if hc == nil {
		hc = NewHTTPClient()
	}

	return &Client{
		BeforeOptions: beforeOptions,
		JSONUnmarshal: json.Unmarshal,
		XMLUnmarshal:  xml.Unmarshal,
		httpClient:    hc,
	}
}
