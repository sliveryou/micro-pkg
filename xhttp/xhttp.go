package xhttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mapping"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/trace"

	"github.com/sliveryou/go-tool/v2/convert"
	"github.com/sliveryou/go-tool/v2/validator"

	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/xhttp/binding"
)

const (
	halfShowLen            = 100
	defaultMultipartMemory = 32 << 20
)

// Response 业务通用响应体
type Response struct {
	TraceID string `json:"trace_id,omitempty" xml:"trace_id,omitempty" example:"a1b2c3d4e5f6g7h8"` // 链路追踪ID
	Code    uint32 `json:"code" xml:"code" example:"0"`                                            // 状态码
	Msg     string `json:"msg" xml:"msg" example:"ok"`                                             // 消息
	Data    any    `json:"data,omitempty" xml:"data,omitempty"`                                    // 数据
}

// GetTraceID 获取链路追踪ID
func GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}

	return ""
}

// WriteHeader 写入自定义响应 header
func WriteHeader(w http.ResponseWriter, err ...error) {
	var ee error
	if len(err) > 0 {
		ee = err[0]
	}

	e, _ := errcode.FromError(ee)
	w.Header().Set(HeaderGWErrorCode, convert.ToString(e.Code))
	w.Header().Set(HeaderGWErrorMessage, url.QueryEscape(e.Msg))
}

// OkJsonCtx 成功 json 响应返回
func OkJsonCtx(ctx context.Context, w http.ResponseWriter, v any) {
	WriteHeader(w)

	httpx.WriteJsonCtx(ctx, w, http.StatusOK, &Response{
		TraceID: GetTraceID(ctx),
		Code:    errcode.CodeOK,
		Msg:     errcode.MsgOK,
		Data:    v,
	})
}

// ErrorCtx 错误响应包装返回
func ErrorCtx(ctx context.Context, w http.ResponseWriter, err error) {
	logx.WithContext(ctx).Errorf("request handle err: %+v", err)

	e, _ := errcode.FromError(err)
	WriteHeader(w, e)

	httpx.WriteJsonCtx(ctx, w, e.HTTPCode, &Response{
		TraceID: GetTraceID(ctx),
		Code:    e.Code,
		Msg:     e.Msg,
		Data:    nil,
	})
}

// Parse 请求体解析
func Parse(r *http.Request, v any) error {
	if err := httpx.Parse(r, v); err != nil {
		logx.WithContext(r.Context()).Errorf("request parse err: %s",
			formatStr(err.Error(), halfShowLen))
		return errcode.ErrInvalidParams
	}

	if err := validator.Verify(v); err != nil {
		return errcode.New(errcode.CodeInvalidParams, err.Error())
	}

	return nil
}

// ParseForm 请求表单解析
func ParseForm(r *http.Request, v any) error {
	if err := binding.Form.Bind(r, v); err != nil {
		logx.WithContext(r.Context()).Errorf("request parse form err: %s",
			formatStr(err.Error(), halfShowLen))
		return errcode.ErrInvalidParams
	}

	if err := validator.Verify(v); err != nil {
		return errcode.New(errcode.CodeInvalidParams, err.Error())
	}

	return nil
}

// ParseJsonBody 解析 json 请求体
func ParseJsonBody(r *http.Request, v any) error {
	if err := mapping.UnmarshalJsonReader(r.Body, v); err != nil {
		logx.WithContext(r.Context()).Errorf("request parse json body err: %s",
			formatStr(err.Error(), halfShowLen))
		return errcode.ErrInvalidParams
	}

	if err := validator.Verify(v); err != nil {
		return errcode.New(errcode.CodeInvalidParams, err.Error())
	}

	return nil
}

// FromFile 请求表单文件获取
func FromFile(r *http.Request, name string) (*multipart.FileHeader, error) {
	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(defaultMultipartMemory); err != nil {
			return nil, err
		}
	}

	f, fh, err := r.FormFile(name)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil, errcode.ErrInvalidParams
		}
		return nil, err
	}
	f.Close()

	return fh, nil
}

// Query 返回给定请求查询参数键的字符串值
func Query(r *http.Request, key string) string {
	value, _ := GetQuery(r, key)

	return value
}

// GetQuery 返回给定请求查询参数键的字符串值并判断其是否存在
func GetQuery(r *http.Request, key string) (string, bool) {
	if values, ok := GetQueryArray(r, key); ok {
		return values[0], ok
	}

	return "", false
}

// QueryArray 返回给定请求查询参数键的字符串切片值
func QueryArray(r *http.Request, key string) []string {
	values, _ := GetQueryArray(r, key)

	return values
}

// GetQueryArray 返回给定请求查询参数键的字符串切片值并判断其是否存在
func GetQueryArray(r *http.Request, key string) ([]string, bool) {
	query := r.URL.Query()
	if values, ok := query[key]; ok && len(values) > 0 {
		return values, true
	}

	return []string{}, false
}

// GetClientIP 获取客户端的IP
func GetClientIP(r *http.Request) string {
	if ip := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); ip != "" {
		return ip
	}

	if ip := strings.TrimSpace(r.Header.Get("X-Real-Ip")); ip != "" {
		return ip
	}

	if addr := strings.TrimSpace(r.Header.Get("X-Appengine-Remote-Addr")); addr != "" {
		return addr
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// GetInternalIP 获取服务端的内部IP
func GetInternalIP() string {
	infs, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, inf := range infs {
		if isEthDown(inf.Flags) || isLoopback(inf.Flags) {
			continue
		}

		addrs, err := inf.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return ""
}

func isEthDown(f net.Flags) bool {
	return f&net.FlagUp != net.FlagUp
}

func isLoopback(f net.Flags) bool {
	return f&net.FlagLoopback == net.FlagLoopback
}

// CopyHTTPRequest 复制请求体
func CopyHTTPRequest(r *http.Request) (*http.Request, error) {
	rClone := r.Clone(context.Background())
	// 克隆请求体
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		rClone.Body = io.NopCloser(bytes.NewReader(body))
	}

	return rClone, nil
}

func formatStr(s string, halfShowLen int) string {
	if length := len(s); length > halfShowLen*2 {
		return s[:halfShowLen] + " ...... " + s[length-halfShowLen-1:]
	}

	return s
}
