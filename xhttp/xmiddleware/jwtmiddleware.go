package xmiddleware

import (
	"net/http"
	"reflect"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/sliveryou/micro-pkg/internal/bizerr"
	"github.com/sliveryou/micro-pkg/jwt"
	"github.com/sliveryou/micro-pkg/xhttp"
)

// ErrInvalidToken Token 错误
var ErrInvalidToken = bizerr.ErrInvalidToken

// -------------------- JWTMiddleware -------------------- //

// JWTMiddleware JWT 认证处理中间件
type JWTMiddleware struct {
	j     *jwt.JWT
	token any
	t     reflect.Type
}

// NewJWTMiddleware 新建 JWT 认证处理中间件
//
// 注意：token 必须为结构体或结构体指针，名称以 json tag 对应的名称与 payloads 进行映射
func NewJWTMiddleware(j *jwt.JWT, token any) (*JWTMiddleware, error) {
	if j == nil || token == nil {
		return nil, errors.New("xmiddleware: illegal jwt middleware config")
	}

	var t reflect.Type
	if jwt.IsStruct(token) {
		t = reflect.ValueOf(token).Type()
	} else if jwt.IsStructPointer(token) {
		t = reflect.ValueOf(token).Elem().Type()
	} else {
		return nil, errors.New("xhttp: check token type err")
	}

	return &JWTMiddleware{j: j, token: token, t: t}, nil
}

// MustNewJWTMiddleware 新建 JWT 认证处理中间件
//
// 注意：token 必须为结构体或结构体指针，名称以 json tag 对应的名称与 payloads 进行映射
func MustNewJWTMiddleware(j *jwt.JWT, token any) *JWTMiddleware {
	m, err := NewJWTMiddleware(j, token)
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
		target := reflect.New(m.t).Interface()

		// 从请求头解析 JWT token，并将其反序列化至指定 token 结构体中
		if err := m.j.ParseTokenFromRequest(r, target); err != nil {
			l.Errorf("jwt middleware parse token err: %v", err)
			xhttp.ErrorCtx(ctx, w, ErrInvalidToken)
			return
		}

		next(w, r.WithContext(jwt.WithCtx(ctx, target)))
	}
}
