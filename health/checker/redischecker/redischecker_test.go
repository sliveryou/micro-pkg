package redischecker

import (
	"context"
	"encoding/json"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var s, _ = miniredis.Run()

func getUpConf() redis.RedisConf {
	s.FlushAll()

	return redis.RedisConf{
		Host: s.Addr(),
		Type: "node",
	}
}

func getDownConf() redis.RedisConf {
	return redis.RedisConf{
		Host:     "unknown host",
		Type:     "node",
		NonBlock: true,
	}
}

func TestNewChecker(t *testing.T) {
	c := NewChecker(getUpConf())
	assert.NotNil(t, c.rds)

	c = NewChecker(getDownConf())
	assert.NotNil(t, c.rds)
}

func TestChecker_Check_Up(t *testing.T) {
	c := NewChecker(getUpConf())
	h := c.Check(context.Background())
	assert.True(t, h.IsUp())

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"UP"}`, string(b))
}

func TestChecker_Check_Down(t *testing.T) {
	c := NewChecker(getDownConf())
	h := c.Check(context.Background())
	assert.True(t, h.IsDown())

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"DOWN"}`, string(b))
}

func TestNewCheckerWithRedis(t *testing.T) {
	kc := getUpConf()
	rds, err := redis.NewRedis(kc)
	require.NoError(t, err)

	c := NewCheckerWithRedis(rds)
	h := c.Check(context.Background())
	assert.True(t, h.IsUp())

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"UP"}`, string(b))
}
