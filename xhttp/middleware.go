package xhttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strings"

	"github.com/go-stack/stack"
	"github.com/zeromicro/go-zero/core/iox"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/sliveryou/micro-pkg/jwt"
	"github.com/sliveryou/micro-pkg/xgrpc"
)

// -------------------- JWTMiddleware -------------------- //

// JWTMiddleware JWT 认证处理中间件
type JWTMiddleware struct {
	j              *jwt.JWT
	token          any
	t              reflect.Type
	errTokenVerify error
}

// NewJWTMiddleware 新建 JWT 认证处理中间件
// 注意：token 必须为结构体指针，名称以 json tag 对应的名称与 payloads 进行映射
func NewJWTMiddleware(j *jwt.JWT, token any, errTokenVerify error) (*JWTMiddleware, error) {
	if err := jwt.CheckTokenType(token); err != nil {
		return nil, err
	}

	return &JWTMiddleware{
		j: j, token: token, t: reflect.ValueOf(token).Elem().Type(), errTokenVerify: errTokenVerify,
	}, nil
}

// MustNewJWTMiddleware 新建 JWT 认证处理中间件
// 注意：token 必须为结构体指针，名称以 json tag 对应的名称与 payloads 进行映射
func MustNewJWTMiddleware(j *jwt.JWT, token any, errTokenVerify error) *JWTMiddleware {
	m, err := NewJWTMiddleware(j, token, errTokenVerify)
	if err != nil {
		panic(err)
	}

	return m
}

// Handle JWT 认证处理
func (m *JWTMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logx.WithContext(ctx)
		target := reflect.New(m.t).Elem().Addr().Interface()

		if err := m.j.ParseTokenFromRequest(r, target); err != nil {
			l.Errorf("jwt middleware parse token from request err: %v", err)
			ErrorCtx(ctx, w, m.errTokenVerify)
			return
		}

		next(w, r.WithContext(jwt.WithCtx(ctx, target)))
	}
}

// -------------------- CorsMiddleware -------------------- //

// CorsMiddleware 跨域请求处理中间件
type CorsMiddleware struct{}

// NewCorsMiddleware 新建跨域请求处理中间件
func NewCorsMiddleware() *CorsMiddleware {
	return &CorsMiddleware{}
}

// Handle 跨域请求处理
func (m *CorsMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setHeader(w)

		// 放行所有 OPTIONS 方法
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 处理请求
		next(w, r)
	}
}

// Handler 跨域请求处理器
func (m *CorsMiddleware) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setHeader(w)

		// 放行所有 OPTIONS 方法
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

// setHeader 设置响应头
func setHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, Authorization, AccessToken, Token, X-Health-Secret")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Access-Control-Allow-Origin, Access-Control-Allow-Headers, X-GW-Error-Code, X-GW-Error-Message")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// -------------------- RecoverMiddleware -------------------- //

// RecoverMiddleware 恐慌捕获恢复处理中间件
type RecoverMiddleware struct{}

// NewRecoverMiddleware 新建恐慌捕获恢复处理中间件
func NewRecoverMiddleware() *RecoverMiddleware {
	return &RecoverMiddleware{}
}

// Handle 恐慌捕获恢复处理
func (m *RecoverMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if cause := recover(); cause != nil {
				logx.WithContext(r.Context()).Errorf("%s%+v [running]:\n%s", dumpRequest(r), cause, getStacks())
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next(w, r)
	}
}

// getStacks 获取调用堆栈信息
func getStacks() string {
	cs := stack.Trace().TrimBelow(stack.Caller(2)).TrimRuntime()
	var b strings.Builder

	for _, c := range cs {
		s := fmt.Sprintf("%+n\n\t%+v", c, c)
		if !strings.Contains(s, "github.com/zeromicro/go-zero") &&
			!strings.Contains(s, "net/http") {
			b.WriteString(s)
			b.WriteString("\n")
		}
	}

	return strings.TrimSpace(b.String())
}

// -------------------- RLogMiddleware -------------------- //

// RLogMiddleware 请求响应日志打印处理中间件
type RLogMiddleware struct{}

// NewRLogMiddleware 新建请求响应日志打印处理中间件
func NewRLogMiddleware() *RLogMiddleware {
	return &RLogMiddleware{}
}

// Handle 请求响应日志打印处理
func (m *RLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dup io.ReadCloser
		writer := NewDetailLoggedResponseWriter(w, r)
		r.Body, dup = iox.DupReadCloser(r.Body)

		next(writer, r)

		r.Body = dup
		logDetails(writer, r)
	}
}

// logDetails 请求响应日志详情打印
func logDetails(writer *DetailLoggedResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%d - %s\n=> %s",
		writer.Writer.Code, httpx.GetRemoteAddr(r), dumpRequest(r)))

	respBuf := writer.Buf.Bytes()
	if len(respBuf) > 0 {
		buf.WriteString(fmt.Sprintf("<= %s", respBuf))
	}

	logx.WithContext(r.Context()).Info(buf.String())
}

// dumpRequest 格式化请求样式
func dumpRequest(req *http.Request) string {
	var dup io.ReadCloser
	req.Body, dup = iox.DupReadCloser(req.Body)

	var b bytes.Buffer
	var err error

	reqURI := req.RequestURI
	if reqURI == "" {
		reqURI = req.URL.RequestURI()
	}

	fmt.Fprintf(&b, "%s %s HTTP/%d.%d\n", req.Method,
		reqURI, req.ProtoMajor, req.ProtoMinor)

	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
	if req.Body != nil {
		var n int64
		var dest io.Writer = &b
		if chunked {
			dest = httputil.NewChunkedWriter(dest)
		}
		n, err = io.Copy(dest, req.Body)
		if closer, ok := dest.(io.Closer); ok && chunked {
			closer.Close()
		}
		if n > 0 {
			io.WriteString(&b, "\n")
		}
	}

	req.Body = dup
	if err != nil {
		return err.Error()
	}

	return b.String()
}

// -------------------- IgnoreRLogMiddleware -------------------- //

// IgnoreRLogMiddleware 忽略 grpc 请求响应日志打印处理中间件
type IgnoreRLogMiddleware struct{}

// NewIgnoreRLogMiddleware 新建忽略 grpc 请求响应日志打印处理中间件
func NewIgnoreRLogMiddleware() *IgnoreRLogMiddleware {
	return &IgnoreRLogMiddleware{}
}

// Handle 忽略 grpc 请求响应日志打印处理
func (m *IgnoreRLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := xgrpc.IgnoreRLog(r.Context())
		next(w, r.WithContext(ctx))
	}
}
