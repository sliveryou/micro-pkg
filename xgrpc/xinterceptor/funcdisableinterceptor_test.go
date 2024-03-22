package xinterceptor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sliveryou/micro-pkg/disabler"
)

func getFuncDisabler() *disabler.FuncDisabler {
	fd := disabler.MustNewFuncDisabler(disabler.Config{
		DisabledRPCs: []string{
			"/auth.Auth/*",
			"/contract.Contract/GetPasses",
			"/file.File/*",
			"/pay.Pay/GetPlan",
		},
	})

	return fd
}

func TestFuncDisableInterceptor(t *testing.T) {
	_, err := FuncDisableInterceptor(getFuncDisabler())(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/contract.Contract/GetPass",
	}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	require.NoError(t, err)

	_, err = FuncDisableInterceptor(getFuncDisabler())(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/auth.Auth/GetPersonalAuth",
	}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})
	require.EqualError(t, err, ErrRPCNotAllowed.Error())
}

func TestFuncDisableStreamInterceptor(t *testing.T) {
	err := FuncDisableStreamInterceptor(getFuncDisabler())(nil, nil, &grpc.StreamServerInfo{
		FullMethod: "/pay.Pay/GetPlans",
	}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	require.NoError(t, err)

	err = FuncDisableStreamInterceptor(getFuncDisabler())(nil, nil, &grpc.StreamServerInfo{
		FullMethod: "/file.File/GetFiles",
	}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	require.EqualError(t, err, ErrRPCNotAllowed.Error())
}
