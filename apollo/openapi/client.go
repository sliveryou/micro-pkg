package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/xhttp"
)

const (
	// ApplicationJSONCharsetUTF8  应用类型：json;charset=UTF-8
	ApplicationJSONCharsetUTF8 = "application/json;charset=UTF-8"

	// DefaultEnv 默认环境
	DefaultEnv = "DEV"
	// DefaultCluster 默认集群
	DefaultCluster = "default"

	httpPrefix  = "http://"
	httpsPrefix = "https://"
)

var (
	// ErrStatusBadRequest 400 - Bad Request
	ErrStatusBadRequest = fmt.Errorf("400 - Bad Request - 客户端传入参数的错误，如操作人不存在，Namespace 不存在等等，客户端需要根据提示信息检查对应的参数是否正确。")
	// ErrStatusUnauthorized 401 - Bad Request
	ErrStatusUnauthorized = fmt.Errorf("401 - Bad Request - 接口传入的 Token 非法或者已过期，客户端需要检查 Token 是否传入正确。")
	// ErrStatusForbidden 403 - Forbidden
	ErrStatusForbidden = fmt.Errorf("403 - Forbidden - 接口要访问的资源未得到授权，比如只授权了对 A 应用下 Namespace 的管理权限，但是却尝试管理 B 应用下的配置。")
	// ErrStatusNotFound 404 - Not Found
	ErrStatusNotFound = fmt.Errorf("404 - Not Found - 接口要访问的资源不存在，一般是 URL 或 URL 的参数错误。")
	// ErrStatusMethodNotAllowed 405 - Method Not Allowed
	ErrStatusMethodNotAllowed = fmt.Errorf("405 - Method Not Allowed - 接口访问的 Method 不正确，比如应该使用 POST 的接口使用了 GET 访问等，客户端需要检查接口访问方式是否正确。")
	// ErrStatusInternalServerError 500 - Internal Server Error
	ErrStatusInternalServerError = fmt.Errorf("500 - Internal Server Error - 其它类型的错误默认都会返回 500，对这类错误如果应用无法根据提示信息找到原因的话，可以找 Apollo 研发团队一起排查问题。")
)

// ClientConfig 默认阿波罗配置中心开放平台客户端配置
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=%e4%ba%8c%e3%80%81-%e7%ac%ac%e4%b8%89%e6%96%b9%e5%ba%94%e7%94%a8%e6%8e%a5%e5%85%a5apollo%e5%bc%80%e6%94%be%e5%b9%b3%e5%8f%b0
type ClientConfig struct {
	PortalAddress    string // 入口地址，一般端口为 8070
	Token            string // 鉴权令牌，须事先在管理平台注册并授权
	DefaultEnv       string // 默认环境，一般为 DEV
	DefaultAppID     string // 默认应用ID
	DefaultCluster   string // 默认集群，一般为 default
	DefaultNamespace string // 默认命名空间
}

// getAPIOption 获取接口可选参数配置
func (c ClientConfig) getAPIOption(opts ...APIOption) apiOption {
	o := apiOption{
		env:       c.DefaultEnv,
		appID:     c.DefaultAppID,
		cluster:   c.DefaultCluster,
		namespace: c.DefaultNamespace,
	}

	for _, opt := range opts {
		opt(&o)
	}

	o.namespace = NormalizeNamespace(o.namespace)

	return o
}

// ClientOption 阿波罗配置中心开放平台客户端可选配置
type ClientOption func(c *client)

// WithHTTPClient 使用配置的 HTTP 客户端
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *client) {
		c.httpClient = hc
	}
}

// NewClient 新建阿波罗配置中心开放平台客户端
func NewClient(config ClientConfig, opts ...ClientOption) (OpenAPI, error) {
	if config.PortalAddress == "" || config.Token == "" || config.DefaultAppID == "" {
		return nil, errors.New("openapi: illegal apollo openapi client configure")
	}

	if config.DefaultEnv == "" {
		config.DefaultEnv = DefaultEnv
	}
	if config.DefaultCluster == "" {
		config.DefaultCluster = DefaultCluster
	}

	config.PortalAddress = NormalizeURL(config.PortalAddress)
	config.DefaultNamespace = NormalizeNamespace(config.DefaultNamespace)

	c := &client{config: config}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = xhttp.NewHTTPClient()
	}

	return c, nil
}

// MustNewClient 新建阿波罗配置中心开放平台客户端
func MustNewClient(config ClientConfig, opts ...ClientOption) OpenAPI {
	c, err := NewClient(config, opts...)
	if err != nil {
		panic(err)
	}

	return c
}

// client 默认阿波罗配置中心开放平台客户端
type client struct {
	config     ClientConfig
	httpClient *http.Client
}

// newRequest 新建请求体
func (c *client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.WithMessagef(err, "http.NewRequestWithContext err, method: %s, url: %s", method, url)
	}

	req.Header.Set(xhttp.HeaderAuthorization, c.config.Token)
	req.Header.Set(xhttp.HeaderContentType, ApplicationJSONCharsetUTF8)

	return req, nil
}

// do 执行请求
func (c *client) do(ctx context.Context, method, url string, request, response interface{}) error {
	var (
		err           error
		reqBody       []byte
		reqBodyReader io.Reader
		req           *http.Request
		respBody      []byte
	)

	if request != nil {
		reqBody, err = json.Marshal(request)
		if err != nil {
			return errors.WithMessage(err, "json.Marshal err")
		}

		reqBodyReader = bytes.NewReader(reqBody)
	}

	req, err = c.newRequest(ctx, method, url, reqBodyReader)
	if err != nil {
		return errors.WithMessage(err, "c.newRequest err")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.WithMessage(err, "c.httpClient.Do err")
	}
	defer resp.Body.Close()

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return errors.WithMessage(err, "io.ReadAll err")
	}

	if resp.StatusCode == http.StatusOK {
		if response != nil {
			return errors.WithMessage(json.Unmarshal(respBody, response), "json.Unmarshal err")
		}

		return nil
	}

	return errors.WithMessagef(parseError(resp.StatusCode),
		"error message: %s", string(respBody))
}

// GetEnvClusters 获取对应应用环境下的所有集群信息（appId 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_321-%e8%8e%b7%e5%8f%96app%e7%9a%84%e7%8e%af%e5%a2%83%ef%bc%8c%e9%9b%86%e7%be%a4%e4%bf%a1%e6%81%af
func (c *client) GetEnvClusters(ctx context.Context, opts ...APIOption) (resp []*EnvClusters, err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/apps/%s/envclusters", c.config.PortalAddress, o.appID)
	err = c.do(ctx, http.MethodGet, url, nil, &resp)
	return
}

// GetNamespaces 获取对应集群下的所有命名空间信息（env、appId 和 cluster 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_325-%e8%8e%b7%e5%8f%96%e9%9b%86%e7%be%a4%e4%b8%8b%e6%89%80%e6%9c%89namespace%e4%bf%a1%e6%81%af%e6%8e%a5%e5%8f%a3
func (c *client) GetNamespaces(ctx context.Context, opts ...APIOption) (resp []*Namespace, err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces",
		c.config.PortalAddress, o.env, o.appID, o.cluster)
	err = c.do(ctx, http.MethodGet, url, nil, &resp)
	return
}

// GetNamespace 获取指定命名空间信息（env、appId、cluster 和 namespace 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_326-%e8%8e%b7%e5%8f%96%e6%9f%90%e4%b8%aanamespace%e4%bf%a1%e6%81%af%e6%8e%a5%e5%8f%a3
func (c *client) GetNamespace(ctx context.Context, opts ...APIOption) (resp *Namespace, err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s",
		c.config.PortalAddress, o.env, o.appID, o.cluster, o.namespace)
	resp = &Namespace{}
	err = c.do(ctx, http.MethodGet, url, nil, resp)
	return
}

// CreateNamespace 创建命名空间信息（appId 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_327-%e5%88%9b%e5%bb%banamespace
func (c *client) CreateNamespace(ctx context.Context, r CreateNamespaceReq, opts ...APIOption) (resp *CreateNamespaceResp, err error) {
	o := c.config.getAPIOption(opts...)
	if r.AppID == "" {
		r.AppID = o.appID
	}
	url := fmt.Sprintf("%s/openapi/v1/apps/%s/appnamespaces",
		c.config.PortalAddress, o.appID)
	resp = &CreateNamespaceResp{}
	err = c.do(ctx, http.MethodPost, url, r, resp)
	return
}

// GetNamespaceLock 获取指定命名空间锁定信息（env、appId、cluster 和 namespace 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_328-%e8%8e%b7%e5%8f%96%e6%9f%90%e4%b8%aanamespace%e5%bd%93%e5%89%8d%e7%bc%96%e8%be%91%e4%ba%ba%e6%8e%a5%e5%8f%a3
func (c *client) GetNamespaceLock(ctx context.Context, opts ...APIOption) (resp *NamespaceLock, err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/lock",
		c.config.PortalAddress, o.env, o.appID, o.cluster, o.namespace)
	resp = &NamespaceLock{}
	err = c.do(ctx, http.MethodGet, url, nil, resp)
	return
}

// AddItem 添加配置信息（env、appId、cluster 和 namespace 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3210-%e6%96%b0%e5%a2%9e%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
func (c *client) AddItem(ctx context.Context, r AddItemReq, opts ...APIOption) (resp *Item, err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items",
		c.config.PortalAddress, o.env, o.appID, o.cluster, o.namespace)
	resp = &Item{}
	err = c.do(ctx, http.MethodPost, url, r, resp)
	return
}

// UpdateItem 更新配置信息（env、appId、cluster 和 namespace 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3211-%e4%bf%ae%e6%94%b9%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
func (c *client) UpdateItem(ctx context.Context, r UpdateItemReq, opts ...APIOption) (err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s",
		c.config.PortalAddress, o.env, o.appID, o.cluster, o.namespace, r.Key)
	err = c.do(ctx, http.MethodPut, url, r, nil)
	return
}

// CreateOrUpdateItem 创建或者更新配置信息（env、appId、cluster 和 namespace 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3211-%e4%bf%ae%e6%94%b9%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
func (c *client) CreateOrUpdateItem(ctx context.Context, r UpdateItemReq, opts ...APIOption) (err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s?createIfNotExists=true",
		c.config.PortalAddress, o.env, o.appID, o.cluster, o.namespace, r.Key)
	err = c.do(ctx, http.MethodPut, url, r, nil)
	return
}

// DeleteItem 删除配置信息（env、appId、cluster 和 namespace 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3212-%e5%88%a0%e9%99%a4%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
func (c *client) DeleteItem(ctx context.Context, r DeleteItemReq, opts ...APIOption) (err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/items/%s?operator=%s",
		c.config.PortalAddress, o.env, o.appID, o.cluster, o.namespace, r.Key, r.Operator)
	err = c.do(ctx, http.MethodDelete, url, nil, nil)
	return
}

// PublishRelease 发布版本配置信息（env、appId、cluster 和 namespace 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3213-%e5%8f%91%e5%b8%83%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
func (c *client) PublishRelease(ctx context.Context, r PublishReleaseReq, opts ...APIOption) (resp *Release, err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/releases",
		c.config.PortalAddress, o.env, o.appID, o.cluster, o.namespace)
	resp = &Release{}
	err = c.do(ctx, http.MethodPost, url, r, resp)
	return
}

// GetRelease 获取版本配置信息（env、appId、cluster 和 namespace 必填）
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3214-%e8%8e%b7%e5%8f%96%e6%9f%90%e4%b8%aanamespace%e5%bd%93%e5%89%8d%e7%94%9f%e6%95%88%e7%9a%84%e5%b7%b2%e5%8f%91%e5%b8%83%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
func (c *client) GetRelease(ctx context.Context, opts ...APIOption) (resp *Release, err error) {
	o := c.config.getAPIOption(opts...)
	url := fmt.Sprintf("%s/openapi/v1/envs/%s/apps/%s/clusters/%s/namespaces/%s/releases/latest",
		c.config.PortalAddress, o.env, o.appID, o.cluster, o.namespace)
	resp = &Release{}
	err = c.do(ctx, http.MethodGet, url, nil, resp)
	return
}

// parseError 解析状态码获取错误信息
// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=%e5%9b%9b%e3%80%81%e9%94%99%e8%af%af%e7%a0%81%e8%af%b4%e6%98%8e
func parseError(status int) error {
	switch status {
	case http.StatusBadRequest:
		return ErrStatusBadRequest
	case http.StatusUnauthorized:
		return ErrStatusUnauthorized
	case http.StatusForbidden:
		return ErrStatusForbidden
	case http.StatusNotFound:
		return ErrStatusNotFound
	case http.StatusMethodNotAllowed:
		return ErrStatusMethodNotAllowed
	case http.StatusInternalServerError:
		return ErrStatusInternalServerError
	default:
		return fmt.Errorf("未定义错误码: %d", status)
	}
}

// NormalizeURL 规范化 url 格式
func NormalizeURL(url string) string {
	if url == "" {
		return ""
	}

	if !strings.HasPrefix(url, httpPrefix) &&
		!strings.HasPrefix(url, httpsPrefix) {
		url = httpPrefix + url
	}

	return strings.TrimSuffix(url, "/")
}

// NormalizeNamespace 规范化命名空间格式
func NormalizeNamespace(ns string) string {
	return strings.TrimSuffix(ns, "."+string(FormatProperties))
}
