package xgrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/sliveryou/micro-pkg/errcode"
)

func TestRLogInterceptor(t *testing.T) {
	_, err := RLogInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.Test/RLogInterceptor",
	}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	require.NoError(t, err)

	_, err = RLogInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.Test/RLogInterceptor",
	}, func(ctx context.Context, req any) (any, error) {
		return nil, errcode.ErrUnexpected
	})
	require.Error(t, err)
}

func TestRLogInterceptor_Ignore(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		rLogKey: ignoreRLogFlag,
	}))

	_, err := RLogInterceptor(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.Test/RLogInterceptor",
	}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	require.NoError(t, err)
}

func TestCrashInterceptor(t *testing.T) {
	_, err := CrashInterceptor(context.Background(), nil, nil,
		func(ctx context.Context, req any) (any, error) {
			panic("mock panic")
		})
	require.Error(t, err)
}

func TestCrashStreamInterceptor(t *testing.T) {
	err := CrashStreamInterceptor(nil, nil, nil, func(
		srv any, stream grpc.ServerStream,
	) error {
		panic("mock panic")
	})
	require.Error(t, err)
}
