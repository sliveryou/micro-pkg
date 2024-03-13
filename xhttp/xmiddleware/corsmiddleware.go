package xmiddleware

import "net/http"

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
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, AccessToken, Token, X-CSRF-Token, X-Health-Secret")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Access-Control-Allow-Origin, Access-Control-Allow-Headers, X-Ca-Error-Code, X-Ca-Error-Message")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
