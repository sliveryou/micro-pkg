package xmiddleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-stack/stack"
	"github.com/zeromicro/go-zero/core/logx"
)

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
