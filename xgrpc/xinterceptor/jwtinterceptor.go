package xinterceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/sliveryou/micro-pkg/jwt"
)

// JWTInterceptor JWT 服务端一元拦截器
func JWTInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return handler(wrapServerContext(ctx), req)
}

// JWTStreamInterceptor JWT 服务端流拦截器
func JWTStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	wss := newWrappedServerStream(ss)
	wss.WrappedContext = wrapServerContext(wss.WrappedContext)

	return handler(srv, wss)
}

// JWTClientInterceptor JWT 客户端一元拦截器
func JWTClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return invoker(wrapClientContext(ctx), method, req, reply, cc, opts...)
}

// JWTStreamClientInterceptor JWT 客户端流拦截器
func JWTStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return streamer(wrapClientContext(ctx), desc, cc, method, opts...)
}

// wrapServerContext 包装服务端上下文
func wrapServerContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	token, ok := jwt.FromMD(md)
	if !ok {
		return ctx
	}

	return context.WithValue(ctx, jwt.TokenKey, token)
}

// wrapClientContext 包装客户端上下文
func wrapClientContext(ctx context.Context) context.Context {
	if tok := ctx.Value(jwt.TokenKey); tok != nil {
		if tokenString, ok := tok.(string); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, string(jwt.TokenKey), tokenString)
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
