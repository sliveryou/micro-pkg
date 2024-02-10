package server

import (
	"context"
	"net"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/sliveryou/micro-pkg/health/checker/compositechecker"
	"github.com/sliveryou/micro-pkg/health/checker/dbchecker"
	"github.com/sliveryou/micro-pkg/health/checker/kvchecker"
	"github.com/sliveryou/micro-pkg/health/checker/redischecker"
	"github.com/sliveryou/micro-pkg/health/checker/rpcchecker"
	"github.com/sliveryou/micro-pkg/health/client"
	"github.com/sliveryou/micro-pkg/xdb"
)

var (
	s1, _ = miniredis.Run()
	s2, _ = miniredis.Run()
)

func TestNewHealthServer(t *testing.T) {
	cc := compositechecker.NewChecker()
	assert.NotPanics(t, func() {
		MustNewHealthServer("test.rpc", cc)
	})
	assert.PanicsWithError(t, "health: illegal health config", func() {
		MustNewHealthServer("test.rpc", nil)
	})
	_, err := NewHealthServer("test.rpc", cc)
	require.NoError(t, err)
	_, err = NewHealthServer("", nil)
	require.EqualError(t, err, "health: illegal health config")
}

func TestHealthServer_Check(t *testing.T) {
	logx.DisableStat()

	dbConf := getDBConf()
	db, mock := xdb.MustNewDBMock(dbConf)
	dbChecker := dbchecker.NewChecker(dbConf.Type, db)
	nodes := getNodes()
	kvChecker := kvchecker.NewCheckerWithNodes(nodes...)
	redisChecker := redischecker.NewCheckerWithRedis(nodes[0])
	rpcChecker := rpcchecker.NewChecker(getCliConf(), zrpc.WithDialOption(grpc.WithContextDialer(dialer(MustNewHealthServer("test.rpc", compositechecker.NewChecker())))))

	cc := compositechecker.NewChecker()
	cc.AddChecker("db", dbChecker)
	cc.AddChecker("kv", kvChecker)
	cc.AddChecker("redis", redisChecker)
	cc.AddChecker("test.rpc", rpcChecker)

	mock.ExpectPrepare("^SELECT 1").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"ok"}).AddRow("1"))
	mock.ExpectPrepare("^SELECT VERSION()").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("5.7.31"))

	hs, err := NewHealthServer("health.rpc", cc)
	require.NoError(t, err)
	assert.NotNil(t, hs)

	hc := client.NewHealthClient(zrpc.MustNewClient(getCliConf(), zrpc.WithDialOption(grpc.WithContextDialer(dialer(hs)))))
	assert.NotNil(t, hc)

	var header metadata.MD
	resp, err := hc.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: ""}, grpc.Header(&header))
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.GetStatus())
	require.Len(t, header[healthHeader], 1)
	assert.Equal(t, `{"db":{"status":"UP","version":"5.7.31"},"kv":{"node0":{"status":"UP"},"node1":{"status":"UP"},"status":"UP"},"redis":{"status":"UP"},"status":"UP","test.rpc":{"status":"UP"}}`, header[healthHeader][0])
}

func getKvConf() kv.KvConf {
	s1.FlushAll()
	s2.FlushAll()

	return []cache.NodeConf{
		{
			RedisConf: redis.RedisConf{
				Host: s1.Addr(),
				Type: "node",
			},
			Weight: 100,
		},
		{
			RedisConf: redis.RedisConf{
				Host: s2.Addr(),
				Type: "node",
			},
			Weight: 100,
		},
	}
}

func getCliConf() zrpc.RpcClientConf {
	return zrpc.RpcClientConf{
		Endpoints: []string{"foo"},
		App:       "foo",
		Token:     "bar",
		Timeout:   1000,
		NonBlock:  true,
	}
}

func getNodes() []*redis.Redis {
	kc := getKvConf()
	nodes := make([]*redis.Redis, 0, len(kc))
	for _, nc := range kc {
		n := redis.MustNewRedis(nc.RedisConf)
		nodes = append(nodes, n)
	}

	return nodes
}

func getDBConf() xdb.Config {
	return xdb.Config{
		Type:     xdb.MySQL,
		Database: "my_test_db",
		LogLevel: xdb.Error,
	}
}

func dialer(srv grpc_health_v1.HealthServer) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, srv)

	go func() {
		if err := server.Serve(listener); err != nil {
			panic(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}
