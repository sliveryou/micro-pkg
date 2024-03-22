package xinterceptor

import (
	"context"

	"google.golang.org/grpc"

	"github.com/sliveryou/micro-pkg/disabler"
	"github.com/sliveryou/micro-pkg/internal/bizerr"
)

// ErrRPCNotAllowed 暂不支持该 RPC 错误
var ErrRPCNotAllowed = bizerr.ErrRPCNotAllowed

// FuncDisableInterceptor 功能禁用服务端一元拦截器
func FuncDisableInterceptor(fd *disabler.FuncDisabler) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !fd.AllowRPC(info.FullMethod) {
			return nil, ErrRPCNotAllowed
		}

		return handler(ctx, req)
	}
}

// FuncDisableStreamInterceptor 功能禁用服务端流拦截器
func FuncDisableStreamInterceptor(fd *disabler.FuncDisabler) grpc.StreamServerInterceptor {
	return func(svr any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !fd.AllowRPC(info.FullMethod) {
			return ErrRPCNotAllowed
		}

		return handler(svr, stream)
	}
}
