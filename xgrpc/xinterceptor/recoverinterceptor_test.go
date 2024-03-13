package xinterceptor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestRecoverInterceptor(t *testing.T) {
	_, err := RecoverInterceptor(context.Background(), nil, nil,
		func(ctx context.Context, req any) (any, error) {
			panic("mock panic")
		})
	require.Error(t, err)
}

func TestRecoverStreamInterceptor(t *testing.T) {
	err := RecoverStreamInterceptor(nil, nil, nil,
		func(srv any, stream grpc.ServerStream) error {
			panic("mock panic")
		})
	require.Error(t, err)
}
