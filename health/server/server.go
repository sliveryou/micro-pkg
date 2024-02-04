package server

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sliveryou/micro-pkg/health"
	"github.com/sliveryou/micro-pkg/health/checker/compositechecker"
)

const healthHeader = "health"

// healthServer 健康检查服务器
type healthServer struct {
	serverName       string                    // 服务名称
	compositeChecker *compositechecker.Checker // 复合应用检查器
	grpc_health_v1.UnimplementedHealthServer
}

// NewHealthServer 新建健康检查服务器
func NewHealthServer(serverName string, compositeChecker *compositechecker.Checker) (grpc_health_v1.HealthServer, error) {
	if serverName == "" {
		serverName = "health"
	}
	if compositeChecker == nil {
		return nil, errors.New("health: illegal health configure")
	}

	return &healthServer{
		serverName:       serverName,
		compositeChecker: compositeChecker,
	}, nil
}

// MustNewHealthServer 新建健康检查服务器
func MustNewHealthServer(serverName string, compositeChecker *compositechecker.Checker) grpc_health_v1.HealthServer {
	s, err := NewHealthServer(serverName, compositeChecker)
	if err != nil {
		panic(err)
	}

	return s
}

// Check 实现了 grpc_health_v1.HealthServer 接口的 Check 方法
func (s *healthServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	l := logx.WithContext(ctx).WithFields(
		logx.Field("service", s.serverName),
		logx.Field("method", "Health.Check"),
	)

	var h health.Health
	if service := in.GetService(); service == "" {
		h = s.compositeChecker.Check(ctx)
	} else {
		h = s.compositeChecker.CheckByName(ctx, service)
	}

	servingStatus := grpc_health_v1.HealthCheckResponse_UNKNOWN
	switch h.GetStatus() {
	case health.Up:
		servingStatus = grpc_health_v1.HealthCheckResponse_SERVING
	case health.Down:
		servingStatus = grpc_health_v1.HealthCheckResponse_NOT_SERVING
	}

	b, err := json.Marshal(h)
	if err != nil {
		l.Errorf("%s service json.Marshal health err: %v", s.serverName, err)
		return nil, status.Errorf(codes.Internal, "%s service json.Marshal health err: %v", s.serverName, err)
	}
	message := string(b)
	l.Infof("%s service check health: %s", s.serverName, message)
	if err := grpc.SendHeader(ctx, metadata.Pairs(healthHeader, message)); err != nil {
		return nil, status.Errorf(codes.Internal, "%s service grpc.SendHeader err: %v", s.serverName, err)
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: servingStatus,
	}, nil
}

// Watch 实现了 grpc_health_v1.HealthServer 接口的 Watch 方法
func (s *healthServer) Watch(in *grpc_health_v1.HealthCheckRequest, w grpc_health_v1.Health_WatchServer) error {
	// 暂不提供 Watch 功能
	return status.Errorf(codes.Unimplemented,
		"%s service Watch() is not implemented", s.serverName)
}
