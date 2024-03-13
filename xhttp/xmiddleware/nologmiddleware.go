package xmiddleware

import (
	"net/http"

	"github.com/sliveryou/micro-pkg/xgrpc/xinterceptor"
)

// -------------------- NoLogMiddleware -------------------- //

// NoLogMiddleware 忽略 grpc 请求响应日志打印处理中间件
type NoLogMiddleware struct{}

// NewNoLogMiddleware 新建忽略 grpc 请求响应日志打印处理中间件
func NewNoLogMiddleware() *NoLogMiddleware {
	return &NoLogMiddleware{}
}

// Handle 忽略 grpc 请求响应日志打印处理
func (m *NoLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := xinterceptor.NoLog(r.Context())
		next(w, r.WithContext(ctx))
	}
}
