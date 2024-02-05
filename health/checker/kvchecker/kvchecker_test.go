package kvchecker

import (
	"context"
	"encoding/json"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func init() {
	logx.ExitOnFatal.Set(false)
}

var (
	s1, _ = miniredis.Run()
	s2, _ = miniredis.Run()
)

func getUpKvConf() kv.KvConf {
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

func getDownKvConf() kv.KvConf {
	s1.FlushAll()

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
				Host:     "unknown host",
				Type:     "node",
				NonBlock: true,
			},
			Weight: 100,
		},
	}
}

func TestNewChecker(t *testing.T) {
	assert.Panics(t, func() {
		NewChecker(kv.KvConf{})
	})

	c := NewChecker(getUpKvConf())
	assert.Len(t, c.nodes, 2)

	c = NewChecker(getDownKvConf())
	assert.Len(t, c.nodes, 2)
}

func TestChecker_Check_Up(t *testing.T) {
	c := NewChecker(getUpKvConf())
	h := c.Check(context.Background())
	assert.True(t, h.IsUp())

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"node0":{"status":"UP"},"node1":{"status":"UP"},"status":"UP"}`, string(b))
}

func TestChecker_Check_Down(t *testing.T) {
	c := NewChecker(getDownKvConf())
	h := c.Check(context.Background())
	assert.True(t, h.IsDown())

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"node0":{"status":"UP"},"node1":{"status":"DOWN"},"status":"DOWN"}`, string(b))
}

func TestNewCheckerWithNodes(t *testing.T) {
	assert.Panics(t, func() {
		NewCheckerWithNodes()
	})

	kc := getUpKvConf()
	nodes := make([]*redis.Redis, 0, len(kc))

	for _, nc := range kc {
		n, err := redis.NewRedis(nc.RedisConf)
		require.NoError(t, err)
		nodes = append(nodes, n)
	}

	c := NewCheckerWithNodes(nodes...)
	h := c.Check(context.Background())
	assert.True(t, h.IsUp())

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"node0":{"status":"UP"},"node1":{"status":"UP"},"status":"UP"}`, string(b))
}
