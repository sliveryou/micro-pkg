package xmiddleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/appsign"
	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/internal/bizerr"
	"github.com/sliveryou/micro-pkg/xhttp"
	"github.com/sliveryou/micro-pkg/xkv"
)

const (
	// KeyPrefixSignNonce 签名随机数缓存 key 前缀
	KeyPrefixSignNonce = "micro.pkg:xhttp.xmiddleware.sign:"

	appErrTime     = 60  // 允许应用误差时间：60s
	signEffTime    = 300 // 签名有效时间：300s
	nonceCacheTime = 300 // 随机数缓存时间：300s
)

var (
	// ErrSignExpired 签名已过期错误
	ErrSignExpired = bizerr.ErrSignExpired
	// ErrNonceExpired 随机数已过期错误
	ErrNonceExpired = bizerr.ErrNonceExpired
)

// GetSecret 密钥查询函数
type GetSecret = func(ctx context.Context, appKey string) (appKeySecret string, err error)

// -------------------- SignMiddleware -------------------- //

// SignMiddleware 签名校验处理中间件
type SignMiddleware struct {
	store     *xkv.Store
	getSecret GetSecret
}

// NewSignMiddleware 新建签名校验处理中间件
func NewSignMiddleware(store *xkv.Store, getSecret GetSecret) (*SignMiddleware, error) {
	if store == nil || getSecret == nil {
		return nil, errors.New("xmiddleware: illegal jwt middleware config")
	}

	return &SignMiddleware{store: store, getSecret: getSecret}, nil
}

// MustNewSignMiddleware 新建签名校验处理中间件
func MustNewSignMiddleware(store *xkv.Store, getSecret GetSecret) *SignMiddleware {
	m, err := NewSignMiddleware(store, getSecret)
	if err != nil {
		panic(err)
	}

	return m
}

// Handle 签名校验处理
func (m *SignMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 从请求中解析应用签名
		appSign, err := appsign.FromRequest(r)
		if err != nil {
			xhttp.ErrorCtx(ctx, w, err)
			return
		}

		// 校验时间戳
		now := time.Now().UnixMilli()
		if appSign.Timestamp-now > appErrTime*1000 ||
			now-appSign.Timestamp > signEffTime*1000 {
			xhttp.ErrorCtx(ctx, w, ErrSignExpired)
			return
		}

		// 获取应用密钥
		secret, err := m.getSecret(ctx, appSign.Key)
		if err != nil {
			xhttp.ErrorCtx(ctx, w, err)
			return
		}

		// 校验签名
		_, ok := appSign.CheckSign(secret)
		if !ok {
			xhttp.ErrorCtx(ctx, w, errcode.New(bizerr.CodeInvalidSign, fmt.Sprintf(
				"签名错误，服务端计算的待签名字符串为 `%s`",
				appSign.StringToSign)))
			return
		}

		// 校验随机数
		key := fmt.Sprintf("%sappkey:%s:nonce:%s", KeyPrefixSignNonce, appSign.Key, appSign.Nonce)
		isExist, err := m.store.ExistsCtx(ctx, key)
		if err != nil {
			xhttp.ErrorCtx(ctx, w, err)
			return
		}
		if isExist {
			xhttp.ErrorCtx(ctx, w, ErrNonceExpired)
			return
		}
		_ = m.store.SetStringCtx(ctx, key, appSign.Nonce, nonceCacheTime)

		next(w, r.WithContext(appsign.CtxWithAppKey(ctx, appSign.Key)))
	}
}
