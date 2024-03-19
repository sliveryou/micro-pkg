package xreq

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

// -------------------- 新建 HTTP 请求 -------------------- //

// New 新建 HTTP 请求
func New(method, url string, options ...Option) (*http.Request, error) {
	request, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		return nil, errors.WithMessage(err, "new http request err")
	}

	return Apply(request, options...)
}

// NewGet 新建 HTTP GET 请求
func NewGet(url string, options ...Option) (*http.Request, error) {
	return New(http.MethodGet, url, options...)
}

// NewPost 新建 HTTP POST 请求
func NewPost(url string, options ...Option) (*http.Request, error) {
	return New(http.MethodPost, url, options...)
}

// NewPut 新建 HTTP PUT 请求
func NewPut(url string, options ...Option) (*http.Request, error) {
	return New(http.MethodPut, url, options...)
}

// NewPatch 新建 HTTP PATCH 请求
func NewPatch(url string, options ...Option) (*http.Request, error) {
	return New(http.MethodPatch, url, options...)
}

// NewDelete 新建 HTTP DELETE 请求
func NewDelete(url string, options ...Option) (*http.Request, error) {
	return New(http.MethodDelete, url, options...)
}

// NewHead 新建 HTTP HEAD 请求
func NewHead(url string, options ...Option) (*http.Request, error) {
	return New(http.MethodHead, url, options...)
}

// NewOptions 新建 HTTP OPTIONS 请求
func NewOptions(url string, options ...Option) (*http.Request, error) {
	return New(http.MethodOptions, url, options...)
}

// -------------------- 发起 HTTP 请求 -------------------- //

// Do 通过 DefaultClient 发起 HTTP 请求
func Do(method, url string, options ...Option) (*Response, error) {
	return DefaultClient.Do(method, OptionCollection{URL(url)}.With(options...)...)
}

// Get 通过 DefaultClient 发起 HTTP GET 请求
func Get(url string, options ...Option) (*Response, error) {
	return Do(http.MethodGet, url, options...)
}

// Post 通过 DefaultClient 发起 HTTP POST 请求
func Post(url string, options ...Option) (*Response, error) {
	return Do(http.MethodPost, url, options...)
}

// Put 通过 DefaultClient 发起 HTTP PUT 请求
func Put(url string, options ...Option) (*Response, error) {
	return Do(http.MethodPut, url, options...)
}

// Patch 通过 DefaultClient 发起 HTTP PATCH 请求
func Patch(url string, options ...Option) (*Response, error) {
	return Do(http.MethodPatch, url, options...)
}

// Delete 通过 DefaultClient 发起 HTTP DELETE 请求
func Delete(url string, options ...Option) (*Response, error) {
	return Do(http.MethodDelete, url, options...)
}

// Head 通过 DefaultClient 发起 HTTP HEAD 请求
func Head(url string, options ...Option) (*Response, error) {
	return Do(http.MethodHead, url, options...)
}

// Options 通过 DefaultClient 发起 HTTP OPTIONS 请求
func Options(url string, options ...Option) (*Response, error) {
	return Do(http.MethodOptions, url, options...)
}
