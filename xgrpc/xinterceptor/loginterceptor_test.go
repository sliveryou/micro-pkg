package xinterceptor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/sliveryou/micro-pkg/errcode"
)

func TestLogInterceptor(t *testing.T) {
	_, err := LogInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.Test/LogInterceptor",
	}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	require.NoError(t, err)

	_, err = LogInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.Test/LogInterceptor",
	}, func(ctx context.Context, req any) (any, error) {
		return nil, errcode.ErrUnexpected
	})
	require.Error(t, err)
}

func TestLogInterceptor_Ignore(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		NoLogKey: NoLogFlag,
	}))

	_, err := LogInterceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.Test/LogInterceptor",
	}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	require.NoError(t, err)
}
