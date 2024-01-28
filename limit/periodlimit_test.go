package limit

import (
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/sliveryou/micro-pkg/xkv"
)

var (
	s1, _ = miniredis.Run()
	s2, _ = miniredis.Run()
)

func runOnPeriodLimit(fn func(pl *PeriodLimit)) {
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

	// 10s 内允许请求 5 次
	fn(MustNewPeriodLimit(10, 5, "period_limit:", store))
}

func TestPeriodLimit_Take(t *testing.T) {
	runOnPeriodLimit(func(pl *PeriodLimit) {
		assert.NotNil(t, pl)

		testKey := "test_key_for_period_limit_take"
		var unknown, allowed, reached, overflowed int

		ticker1 := time.NewTicker(500 * time.Microsecond)
		defer ticker1.Stop()
		ticker2 := time.NewTicker(5 * time.Second)
		defer ticker2.Stop()

	out:
		for {
			select {
			case <-ticker1.C:
				code, _ := pl.Take(testKey, WithQuota(10))
				switch code {
				case Unknown:
					unknown++
				case Allowed:
					allowed++
				case Reached:
					reached++
				case Overflowed:
					overflowed++
				}
			case <-ticker2.C:
				break out
			}
		}

		assert.Equal(t, 9, allowed)
		assert.Equal(t, 1, reached)
		t.Log(unknown, allowed, reached, overflowed)
	})
}

func TestPeriodLimit_Get(t *testing.T) {
	runOnPeriodLimit(func(pl *PeriodLimit) {
		assert.NotNil(t, pl)

		testKey := "test_key_for_period_limit_get"
		var lessThanZero, equalZero, greaterThanZero int

		ticker1 := time.NewTicker(500 * time.Microsecond)
		defer ticker1.Stop()
		ticker2 := time.NewTicker(5 * time.Second)
		defer ticker2.Stop()

	out:
		for {
			select {
			case <-ticker1.C:
				remain, _ := pl.Get(testKey)
				if remain > 0 {
					greaterThanZero++
				} else if remain == 0 {
					equalZero++
				} else {
					lessThanZero++
				}
			case <-ticker2.C:
				break out
			}
		}

		assert.Equal(t, 1, equalZero)
		assert.Equal(t, 4, greaterThanZero)
		t.Log(lessThanZero, equalZero, greaterThanZero)
	})
}

func TestPeriodLimit_Allow(t *testing.T) {
	runOnPeriodLimit(func(pl *PeriodLimit) {
		assert.NotNil(t, pl)

		testKey := "test_key_for_period_limit_allow"
		var allowed, notAllowed int

		ticker1 := time.NewTicker(500 * time.Microsecond)
		defer ticker1.Stop()
		ticker2 := time.NewTicker(5 * time.Second)
		defer ticker2.Stop()

	out:
		for {
			select {
			case <-ticker1.C:
				ok, _ := pl.Allow(testKey)
				if ok {
					allowed++
				} else {
					notAllowed++
				}
			case <-ticker2.C:
				break out
			}
		}

		assert.Equal(t, 5, allowed)
		t.Log(allowed, notAllowed)
	})
}
