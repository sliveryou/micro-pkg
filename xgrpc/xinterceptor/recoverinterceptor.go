package xinterceptor

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-stack/stack"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoverInterceptor 恐慌捕获恢复服务端一元拦截器
func RecoverInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if cause := recover(); cause != nil {
			err = toPanicError(ctx, cause)
		}
	}()

	return handler(ctx, req)
}

// RecoverStreamInterceptor 恐慌捕获恢复服务端流拦截器
func RecoverStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if cause := recover(); cause != nil {
			ctx := context.Background()
			if ss != nil {
				ctx = ss.Context()
			}
			err = toPanicError(ctx, cause)
		}
	}()

	return handler(srv, ss)
}

// toPanicError 恐慌错误转换
func toPanicError(ctx context.Context, cause any) error {
	logx.WithContext(ctx).Errorf("%+v [running]:\n%s", cause, getStacks())
	return status.Errorf(codes.Internal, "panic: %v", cause)
}

// getStacks 获取调用堆栈信息
func getStacks() string {
	cs := stack.Trace().TrimBelow(stack.Caller(3)).TrimRuntime()
	var b strings.Builder

	for _, c := range cs {
		s := fmt.Sprintf("%+n\n\t%+v", c, c)
		if !strings.Contains(s, "github.com/zeromicro/go-zero") &&
			!strings.Contains(s, "google.golang.org/grpc") {
			b.WriteString(s)
			b.WriteString("\n")
		}
	}

	return strings.TrimSpace(b.String())
}
