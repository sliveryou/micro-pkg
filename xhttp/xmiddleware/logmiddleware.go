package xmiddleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/zeromicro/go-zero/core/iox"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/sliveryou/micro-pkg/xhttp"
)

// -------------------- LogMiddleware -------------------- //

// LogMiddleware 请求响应日志打印处理中间件
type LogMiddleware struct{}

// NewLogMiddleware 新建请求响应日志打印处理中间件
func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

// Handle 请求响应日志打印处理
func (m *LogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dup io.ReadCloser
		writer := xhttp.NewDetailLoggedResponseWriter(w)
		r.Body, dup = iox.DupReadCloser(r.Body)

		next(writer, r)

		r.Body = dup
		logDetails(writer, r)
	}
}

// logDetails 请求响应日志详情打印
func logDetails(writer *xhttp.DetailLoggedResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	code := writer.W.Code
	logger := logx.WithContext(r.Context())

	buf.WriteString(fmt.Sprintf("%d - %s\n=> %s",
		code, httpx.GetRemoteAddr(r), dumpRequest(r)))

	respBuf := writer.Buf.Bytes()
	if len(respBuf) > 0 {
		buf.WriteString(fmt.Sprintf("<= %s", respBuf))
	}

	if code < http.StatusInternalServerError {
		logger.Info(buf.String())
	} else {
		logger.Error(buf.String())
	}
}

// dumpRequest 格式化请求样式
func dumpRequest(r *http.Request) string {
	reqContent, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err.Error()
	}

	return string(reqContent)
}
