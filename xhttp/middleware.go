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
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/iox"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/sliveryou/micro-pkg/disabler"
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
//
// 注意：token 必须为结构体指针，名称以 json tag 对应的名称与 payloads 进行映射
func NewJWTMiddleware(j *jwt.JWT, token any, errTokenVerify error) (*JWTMiddleware, error) {
	if j == nil || token == nil || errTokenVerify == nil {
		return nil, errors.New("xhttp: illegal jwt middleware config")
	}
	if err := jwt.CheckTokenType(token); err != nil {
		return nil, errors.WithMessage(err, "xhttp: check token type err")
	}

	return &JWTMiddleware{
		j: j, token: token, t: reflect.ValueOf(token).Elem().Type(), errTokenVerify: errTokenVerify,
	}, nil
}

// MustNewJWTMiddleware 新建 JWT 认证处理中间件
//
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

		// 从请求头解析 JWT token，并将其反序列化至指定 token 结构体中
		if err := m.j.ParseTokenFromRequest(r, target); err != nil {
			l.Errorf("jwt middleware parse token err: %v", err)
			ErrorCtx(ctx, w, m.errTokenVerify)
			return
		}

		next(w, r.WithContext(jwt.WithCtx(ctx, target)))
	}
}

// -------------------- FuncDisable -------------------- //

// FuncDisableMiddleware 功能禁用处理中间件
type FuncDisableMiddleware struct {
	fd            *disabler.FuncDisabler
	routePrefix   string
	errNotAllowed error
}

// NewFuncDisableMiddleware 新建功能禁用处理中间件
func NewFuncDisableMiddleware(fd *disabler.FuncDisabler, routePrefix string, errNotAllowed error) (*FuncDisableMiddleware, error) {
	if fd == nil || errNotAllowed == nil {
		return nil, errors.New("xhttp: illegal func disable middleware config")
	}

	return &FuncDisableMiddleware{
		fd: fd, routePrefix: routePrefix, errNotAllowed: errNotAllowed,
	}, nil
}

// MustNewFuncDisableMiddleware 新建功能禁用处理中间件
func MustNewFuncDisableMiddleware(fd *disabler.FuncDisabler, routePrefix string, errNotAllowed error) *FuncDisableMiddleware {
	m, err := NewFuncDisableMiddleware(fd, routePrefix, errNotAllowed)
	if err != nil {
		panic(err)
	}

	return m
}

// Handle 功能禁用处理
func (m *FuncDisableMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 当前接口的请求方法
		method := r.Method
		// 当前接口的请求路径
		path := r.URL.Path
		// 去除路径前缀
		api := strings.TrimPrefix(path, m.routePrefix)

		if !m.fd.AllowAPI(method, api) {
			ErrorCtx(r.Context(), w, m.errNotAllowed)
			return
		}

		next(w, r)
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
		writer := NewDetailLoggedResponseWriter(w)
		r.Body, dup = iox.DupReadCloser(r.Body)

		next(writer, r)

		r.Body = dup
		logDetails(writer, r)
	}
}

// logDetails 请求响应日志详情打印
func logDetails(writer *DetailLoggedResponseWriter, r *http.Request) {
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
