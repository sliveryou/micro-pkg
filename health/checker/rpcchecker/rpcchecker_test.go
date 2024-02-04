package rpcchecker

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func init() {
	logx.Disable()
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, &healthServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func getConf() zrpc.RpcClientConf {
	return zrpc.RpcClientConf{
		Endpoints: []string{"foo"},
		App:       "foo",
		Token:     "bar",
		Timeout:   1000,
		NonBlock:  true,
	}
}

func TestNewChecker(t *testing.T) {
	c := NewChecker(getConf(), zrpc.WithDialOption(grpc.WithContextDialer(dialer())))
	h := c.Check(context.Background())
	assert.True(t, h.IsUp())
	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"node":"node1","status":"UP","version":"v1.0.0"}`, string(b))
}

func TestNewCheckerWithClient(t *testing.T) {
	hc := zrpc.MustNewClient(getConf(), zrpc.WithDialOption(grpc.WithContextDialer(dialer())))
	c := NewCheckerWithClient(hc)
	h := c.Check(context.Background())
	assert.True(t, h.IsUp())
	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"node":"node1","status":"UP","version":"v1.0.0"}`, string(b))
}

type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *healthServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if err := grpc.SendHeader(ctx, metadata.Pairs("health", `{"status":"UP","node":"node1","version":"v1.0.0"}`)); err != nil {
		return nil, status.Errorf(codes.Internal, "grpc.SendHeader err: %v", err)
	}
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *healthServer) Watch(in *grpc_health_v1.HealthCheckRequest, w grpc_health_v1.Health_WatchServer) error {
	return nil
}
