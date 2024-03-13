package xinterceptor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/sliveryou/micro-pkg/errcode"
)

const (
	// NoLogKey 忽略请求响应日志打印 key
	NoLogKey = "X-No-Log"
	// NoLogFlag 忽略请求响应日志打印 flag
	NoLogFlag = "true"
)

// NoLog 忽略 grpc 服务端的请求响应日志打印
func NoLog(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, NoLogKey, NoLogFlag)
}

// LogInterceptor 请求响应日志打印服务端一元拦截器
func LogInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if data := md.Get(NoLogKey); len(data) > 0 && data[0] == NoLogFlag {
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
			buf.WriteString("\n" + string(reqContent))
		}
	}

	if resp != nil {
		respContent, err := json.Marshal(resp)
		if err == nil && len(respContent) > 0 {
			buf.WriteString("\n<= " + string(respContent))
		}
	}

	logx.WithContext(ctx).Info(buf.String())

	return resp, err
}
