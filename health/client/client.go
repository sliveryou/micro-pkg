package client

import (
	"context"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// 类型定义
type (
	defaultHealthClient struct {
		cli zrpc.Client
	}
)

// NewHealthClient 新建 Health 客户端
func NewHealthClient(cli zrpc.Client) grpc_health_v1.HealthClient {
	return &defaultHealthClient{
		cli: cli,
	}
}

// Check if the requested service is unknown, the call will fail with status.
func (m *defaultHealthClient) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest, opts ...grpc.CallOption) (*grpc_health_v1.HealthCheckResponse, error) {
	client := grpc_health_v1.NewHealthClient(m.cli.Conn())
	return client.Check(ctx, in, opts...)
}

// Watch performs a watch for the serving status of the requested service.
func (m *defaultHealthClient) Watch(ctx context.Context, in *grpc_health_v1.HealthCheckRequest, opts ...grpc.CallOption) (grpc_health_v1.Health_WatchClient, error) {
	client := grpc_health_v1.NewHealthClient(m.cli.Conn())
	return client.Watch(ctx, in, opts...)
}
