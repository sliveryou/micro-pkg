package limit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/sliveryou/go-tool/v2/timex"

	"github.com/sliveryou/micro-pkg/xkv"
)

var (
	rate     = 10
	capacity = 50
)

func runOnTokenLimit(fn func(tl *TokenLimit)) {
	s1.FlushAll()
	s2.FlushAll()

	store := xkv.NewStore([]cache.NodeConf{
		{
			RedisConf: redis.RedisConf{
				Host: s1.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
		{
			RedisConf: redis.RedisConf{
				Host: s2.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
	})

	// 创建一个最大容量为 50，每秒生成 10 个令牌的令牌桶限流器
	fn(MustNewTokenLimit(rate, capacity, "token_limit_test", store))
}

func TestTokenLimit_Allow(t *testing.T) {
	runOnTokenLimit(func(tl *TokenLimit) {
		assert.NotNil(t, tl)

		var allowed int
		for i := 0; i < 100; i++ {
			time.Sleep(time.Second / time.Duration(100))
			isAllowed, err := tl.Allow()
			require.NoError(t, err)
			if isAllowed {
				allowed++
			}
		}

		t.Log(allowed)
		assert.GreaterOrEqual(t, allowed, rate+capacity)
	})
}

func TestTokenLimit_AllowN(t *testing.T) {
	runOnTokenLimit(func(tl *TokenLimit) {
		assert.NotNil(t, tl)

		var allowed int
		for i := 0; i < 100; i++ {
			isAllowed, err := tl.AllowN(timex.Now(), 1)
			require.NoError(t, err)
			if isAllowed {
				allowed++
			}
		}

		t.Log(allowed)
		assert.GreaterOrEqual(t, allowed, capacity)
	})
}
