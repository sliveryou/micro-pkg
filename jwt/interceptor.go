package jwt

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
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

// TokenInterceptor 默认令牌服务端一元拦截器
func TokenInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return handler(wrapServerContext(ctx), req)
}

// TokenStreamInterceptor 默认令牌服务端流拦截器
func TokenStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	wss := newWrappedServerStream(ss)
	wss.WrappedContext = wrapServerContext(wss.WrappedContext)

	return handler(srv, wss)
}

// TokenClientInterceptor 默认令牌客户端一元拦截器
func TokenClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return invoker(wrapClientContext(ctx), method, req, reply, cc, opts...)
}

// TokenStreamClientInterceptor 默认令牌客户端流拦截器
func TokenStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return streamer(wrapClientContext(ctx), desc, cc, method, opts...)
}

// wrapServerContext 包装服务端上下文
func wrapServerContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	token, ok := FromMD(md)
	if !ok {
		return ctx
	}

	return context.WithValue(ctx, TokenKey, token)
}

// wrapClientContext 包装客户端上下文
func wrapClientContext(ctx context.Context) context.Context {
	if tok := ctx.Value(TokenKey); tok != nil {
		if tokenString, ok := tok.(string); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, string(TokenKey), tokenString)
		}
	}

	return ctx
}

// wrappedServerStream 包装后的服务端流对象
type wrappedServerStream struct {
	grpc.ServerStream
	WrappedContext context.Context
}

// newWrappedServerStream 新建包装后的服务端流对象
func newWrappedServerStream(ss grpc.ServerStream) *wrappedServerStream {
	if existing, ok := ss.(*wrappedServerStream); ok {
		return existing
	}
	return &wrappedServerStream{ServerStream: ss, WrappedContext: ss.Context()}
}

// Context 返回包装后的服务端流对象的上下文信息
func (w *wrappedServerStream) Context() context.Context {
	return w.WrappedContext
}
