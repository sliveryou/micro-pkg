package xreq

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sliveryou/go-tool/v2/timex"

	"github.com/sliveryou/micro-pkg/xhttp"
)

// Response HTTP 拓展响应
type Response struct {
	RawResponse *http.Response
	body        []byte
	size        int64
	receivedAt  time.Time
	client      *Client
}

// IsSuccess 判断响应状态码是否为成功状态（code >= 200 and <= 299）
func (r *Response) IsSuccess() bool {
	return r.StatusCode() > 199 && r.StatusCode() < 300
}

// IsError 判断响应状态码是否为错误状态（code >= 400）
func (r *Response) IsError() bool {
	return r.StatusCode() > 399
}

// JSONUnmarshal 将 JSON 形式的响应体内容使用 Client.JSONUnmarshal 反序列化到指定对象中
func (r *Response) JSONUnmarshal(v any) error {
	return r.client.JSONUnmarshal(r.Bytes(), v)
}

// XMLUnmarshal 将 XML 形式的响应体内容使用 Client.XMLUnmarshal 反序列化到指定对象中
func (r *Response) XMLUnmarshal(v any) error {
	return r.client.XMLUnmarshal(r.Bytes(), v)
}

// Unmarshal 根据响应头部 "Content-Type" 的值将响应体内容使用特定 Unmarshaler
// 函数反序列化到指定对象中
func (r *Response) Unmarshal(v any) error {
	if ct := r.ContentType(); strings.Contains(ct, "json") {
		return errors.WithMessage(r.JSONUnmarshal(v), "json unmarshal err")
	} else if strings.Contains(ct, "xml") {
		return errors.WithMessage(r.XMLUnmarshal(v), "xml unmarshal err")
	}

	return errors.WithMessage(r.JSONUnmarshal(v), "json unmarshal err")
}

// Bytes 返回 []byte 形式的响应体内容
func (r *Response) Bytes() []byte {
	return r.body
}

// String 返回 string 形式的响应体内容
func (r *Response) String() string {
	if len(r.body) == 0 {
		return ""
	}

	return strings.TrimSpace(string(r.body))
}

// Size 返回响应体大小
func (r *Response) Size() int64 {
	return r.size
}

// ReceivedAt 响应接收时间
func (r *Response) ReceivedAt() time.Time {
	return r.receivedAt
}

// RawBody 返回原始响应体
func (r *Response) RawBody() io.ReadCloser {
	if r.RawResponse == nil {
		return nil
	}

	return r.RawResponse.Body
}

// Status 返回响应状态
//
//	Example: 200 OK
func (r *Response) Status() string {
	if r.RawResponse == nil {
		return ""
	}

	return r.RawResponse.Status
}

// StatusCode 返回响应状态码
//
//	Example: 200
func (r *Response) StatusCode() int {
	if r.RawResponse == nil {
		return 0
	}

	return r.RawResponse.StatusCode
}

// Proto 返回响应协议
func (r *Response) Proto() string {
	if r.RawResponse == nil {
		return ""
	}

	return r.RawResponse.Proto
}

// ContentType 返回响应头部 "Content-Type" 的值
func (r *Response) ContentType() string {
	return r.Header().Get(xhttp.HeaderContentType)
}

// Header 返回响应头部
func (r *Response) Header() http.Header {
	if r.RawResponse == nil {
		return http.Header{}
	}

	return r.RawResponse.Header
}

// GetAllHeadersString 返回 string 形式的所有响应头部内容
func (r *Response) GetAllHeadersString() string {
	if r.RawResponse == nil || r.RawResponse.Header == nil {
		return ""
	}

	buf := new(bytes.Buffer)
	_ = r.RawResponse.Header.Write(buf)

	return buf.String()
}

// Cookies 返回响应 cookie 列表
func (r *Response) Cookies() []*http.Cookie {
	if r.RawResponse == nil {
		return make([]*http.Cookie, 0)
	}

	return r.RawResponse.Cookies()
}

// setReceivedAt 设置接收时间
func (r *Response) setReceivedAt() {
	r.receivedAt = timex.Now()
}
