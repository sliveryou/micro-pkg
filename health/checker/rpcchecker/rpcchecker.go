package rpcchecker

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	"github.com/sliveryou/micro-pkg/health"
	"github.com/sliveryou/micro-pkg/health/client"
)

var _ health.Checker = (*Checker)(nil)

const healthHeader = "health"

// Checker 服务检查器结构详情
type Checker struct {
	hc grpc_health_v1.HealthClient
}

// NewChecker 新建服务检查器
func NewChecker(cc zrpc.RpcClientConf, options ...zrpc.ClientOption) *Checker {
	return &Checker{hc: client.NewHealthClient(zrpc.MustNewClient(cc, options...))}
}

// NewCheckerWithClient 通过已有客户端新建服务检查器
func NewCheckerWithClient(cli zrpc.Client) *Checker {
	return &Checker{hc: client.NewHealthClient(cli)}
}

// Check 检查服务健康情况
func (c *Checker) Check(ctx context.Context) health.Health {
	h := health.NewHealth()

	var header metadata.MD
	resp, err := c.hc.Check(ctx, &grpc_health_v1.HealthCheckRequest{}, grpc.Header(&header))
	if err != nil {
		h.Down().AddInfo("error", err.Error())
		return h
	} else if len(header[healthHeader]) > 0 {
		healthBytes := []byte(header[healthHeader][0])
		if err := json.Unmarshal(healthBytes, &h); err == nil {
			return h
		}
	}

	switch resp.GetStatus() {
	case grpc_health_v1.HealthCheckResponse_UNKNOWN:
		h.Unknown()
	case grpc_health_v1.HealthCheckResponse_SERVING:
		h.Up()
	default:
		h.Down()
	}

	return h
}