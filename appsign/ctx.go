package appsign

import "context"

const (
	// ContextAppKey AppKey 上下文
	ContextAppKey = contextKey("X-AppKey")
)

// contextKey 上下文 key 类型
type contextKey string

// String 实现序列化字符串方法
func (c contextKey) String() string {
	return "app sign context key: " + string(c)
}

// CtxWithAppKey 将指定 AppKey 关联到 context 中
func CtxWithAppKey(ctx context.Context, appkey string) context.Context {
	return context.WithValue(ctx, ContextAppKey, appkey)
}

// AppKeyFromCtx 从 context 获取 AppKey
func AppKeyFromCtx(ctx context.Context) string {
	val := ctx.Value(ContextAppKey)
	if appkey, ok := val.(string); ok {
		return appkey
	}

	return ""
}
