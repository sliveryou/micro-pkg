package xmiddleware

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/disabler"
	"github.com/sliveryou/micro-pkg/internal/bizerr"
	"github.com/sliveryou/micro-pkg/xhttp"
)

// ErrAPINotAllowed 暂不支持该 API 错误
var ErrAPINotAllowed = bizerr.ErrAPINotAllowed

// -------------------- FuncDisable -------------------- //

// FuncDisableMiddleware 功能禁用处理中间件
type FuncDisableMiddleware struct {
	fd          *disabler.FuncDisabler
	routePrefix string
}

// NewFuncDisableMiddleware 新建功能禁用处理中间件
func NewFuncDisableMiddleware(fd *disabler.FuncDisabler, routePrefix string) (*FuncDisableMiddleware, error) {
	if fd == nil {
		return nil, errors.New("xmiddleware: illegal func disable middleware config")
	}

	return &FuncDisableMiddleware{fd: fd, routePrefix: routePrefix}, nil
}

// MustNewFuncDisableMiddleware 新建功能禁用处理中间件
func MustNewFuncDisableMiddleware(fd *disabler.FuncDisabler, routePrefix string) *FuncDisableMiddleware {
	m, err := NewFuncDisableMiddleware(fd, routePrefix)
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
			xhttp.ErrorCtx(r.Context(), w, ErrAPINotAllowed)
			return
		}

		next(w, r)
	}
}
