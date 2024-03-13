package xinterceptor

import (
	"context"

	"google.golang.org/grpc"

	"github.com/sliveryou/micro-pkg/disabler"
)

// FuncDisableInterceptor 功能禁用服务端一元拦截器
func FuncDisableInterceptor(fd *disabler.FuncDisabler, errNotAllowed error) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !fd.AllowRPC(info.FullMethod) {
			return nil, errNotAllowed
		}

		return handler(ctx, req)
	}
}

// FuncDisableStreamInterceptor 功能禁用服务端流拦截器
func FuncDisableStreamInterceptor(fd *disabler.FuncDisabler, errNotAllowed error) grpc.StreamServerInterceptor {
	return func(svr any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !fd.AllowRPC(info.FullMethod) {
			return errNotAllowed
		}

		return handler(svr, stream)
	}
}
