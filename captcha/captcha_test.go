package captcha

import (
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/sliveryou/micro-pkg/xkv"
)

var (
	s1, _ = miniredis.Run()
	s2, _ = miniredis.Run()
)

func getStore() *xkv.Store {
	s1.FlushAll()
	s2.FlushAll()

	return xkv.NewStore([]cache.NodeConf{
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
	})
}

func TestNewCaptcha(t *testing.T) {
	c, err := NewCaptcha(Config{}, getStore())
	require.NoError(t, err)
	assert.NotNil(t, c)

	assert.PanicsWithError(t, "captcha: illegal captcha config", func() {
		MustNewCaptcha(Config{}, nil)
	})
}

func TestCaptcha_Generate(t *testing.T) {
	c, err := NewCaptcha(Config{}, getStore())
	require.NoError(t, err)

	id, b64s, answer, err := c.Generate()
	require.NoError(t, err)
	t.Log(id, answer, b64s)

	value := c.captcha.Store.Get(id, true)
	t.Log(value)
	assert.Equal(t, answer, value)

	value = c.captcha.Store.Get(id, false)
	assert.Empty(t, value)
}

func TestCaptcha_Verify(t *testing.T) {
	c, err := NewCaptcha(Config{}, getStore())
	require.NoError(t, err)

	id, b64s, answer, err := c.Generate()
	require.NoError(t, err)
	t.Log(id, answer, b64s)

	value := c.captcha.Store.Get(id, false)
	t.Log(value)
	assert.Equal(t, answer, value)

	ok := c.Verify(id, "unknown value", false)
	assert.False(t, ok)

	ok = c.Verify(id, value, true)
	assert.True(t, ok)

	value = c.captcha.Store.Get(id, false)
	assert.Empty(t, value)
}
