package xgrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-stack/stack"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/sliveryou/micro-pkg/errcode"
)

const (
	// rLogKey 忽略请求响应日志打印 key
	rLogKey = "X-Ignore-RLog"
	// ignoreRLogFlag 忽略请求响应日志打印 flag
	ignoreRLogFlag = "true"
)

// IgnoreRLog 忽略 grpc 服务端的请求响应日志打印
func IgnoreRLog(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, rLogKey, ignoreRLogFlag)
}

// RLogInterceptor 请求响应日志打印服务端一元拦截器
func RLogInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if data := md.Get(rLogKey); len(data) > 0 && data[0] == ignoreRLogFlag {
			return resp, err
		}
	}

	e, _ := errcode.FromError(err)
	a := "unknown addr"
	p, ok := peer.FromContext(ctx)
	if ok {
		a = p.Addr.String()
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%d - %s\n=> %s", e.Code, a, info.FullMethod))

	if req != nil {
		reqContent, err := json.Marshal(req)
		if err == nil && len(reqContent) > 0 {
			buf.WriteString(fmt.Sprintf("\n%s", string(reqContent)))
		}
	}

	if resp != nil {
		respContent, err := json.Marshal(resp)
		if err == nil && len(respContent) > 0 {
			buf.WriteString(fmt.Sprintf("\n<= %s", string(respContent)))
		}
	}

	logx.WithContext(ctx).Info(buf.String())

	return resp, err
}

// CrashInterceptor 恐慌捕获恢复服务端一元拦截器
func CrashInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if cause := recover(); cause != nil {
			err = toPanicError(ctx, cause)
		}
	}()

	return handler(ctx, req)
}

// CrashStreamInterceptor 恐慌捕获恢复服务端流拦截器
func CrashStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
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
