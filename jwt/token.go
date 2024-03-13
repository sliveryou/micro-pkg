package jwt

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/metadata"
)

const (
	// TokenKey 用户令牌 key
	TokenKey = contextKey("X-Token")
)

// contextKey 上下文 key 类型
type contextKey string

// String 实现序列化字符串方法
func (c contextKey) String() string {
	return "jwt context key: " + string(c)
}

// WithCtx 将指定 token 令牌数据关联到 context 中
func WithCtx(ctx context.Context, token any) context.Context {
	if token != nil {
		if tokenBytes, err := json.Marshal(token); err == nil {
			return context.WithValue(ctx, TokenKey, string(tokenBytes))
		}
	}

	return ctx
}

// ReadCtx 从 context 获取令牌数据，并将其反序列化至指定 token 中
//
// 注意：token 必须为指针类型
func ReadCtx(ctx context.Context, token any) error {
	if tok := ctx.Value(TokenKey); tok != nil {
		if tokenString, ok := tok.(string); ok {
			return json.Unmarshal([]byte(tokenString), token)
		}
	}

	return errNoTokenInCtx
}

// FromMD 从 grpc metadata 获取令牌数据
func FromMD(md metadata.MD) (string, bool) {
	if md != nil {
		if ts := md.Get(string(TokenKey)); len(ts) > 0 {
			return ts[0], true
		}
	}

	return "", true
}
