package xhttp

import (
	"net"
	"net/http"
	"time"
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
