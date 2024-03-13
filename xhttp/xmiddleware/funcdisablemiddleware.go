package xmiddleware

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/disabler"
	"github.com/sliveryou/micro-pkg/xhttp"
)

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
			xhttp.ErrorCtx(r.Context(), w, m.errNotAllowed)
			return
		}

		next(w, r)
	}
}
