package kvchecker

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/sliveryou/go-tool/v2/convert"

	"github.com/sliveryou/micro-pkg/health"
)

var _ health.Checker = (*Checker)(nil)

// Checker 键值集群检查器结构详情
type Checker struct {
	nodes []*redis.Redis
}

// NewChecker 新建键值集群检查器
func NewChecker(kc kv.KvConf, opts ...redis.Option) *Checker {
	if len(kc) == 0 || cache.TotalWeights(kc) <= 0 {
		logx.Must(errors.New("there are no cache nodes"))
	}

	c := &Checker{nodes: make([]*redis.Redis, 0, len(kc))}
	for _, nc := range kc {
		n := redis.MustNewRedis(nc.RedisConf, opts...)
		c.nodes = append(c.nodes, n)
	}

	return c
}

// NewCheckerWithNodes 通过已有节点新建键值集群检查器
func NewCheckerWithNodes(nodes ...*redis.Redis) *Checker {
	if len(nodes) == 0 {
		logx.Must(errors.New("there are no cache nodes"))
	}

	c := &Checker{nodes: make([]*redis.Redis, 0, len(nodes))}
	for _, node := range nodes {
		if node != nil {
			c.nodes = append(c.nodes, node)
		}
	}

	return c
}

// Check 检查键值集群健康情况
func (c *Checker) Check(ctx context.Context) health.Health {
	h := health.NewHealth()
	h.Up()

	for i, n := range c.nodes {
		nh := health.NewHealth()
		nh.Down()

		if n.PingCtx(ctx) {
			nh.Up()
		} else if !h.IsDown() {
			h.Down()
		}

		h.AddInfo("node"+convert.ToString(i), nh)
	}

	return h
}
