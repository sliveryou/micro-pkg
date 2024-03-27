package redischecker

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/sliveryou/micro-pkg/health"
)

var _ health.Checker = (*Checker)(nil)

// Checker redis 检查器
type Checker struct {
	rds *redis.Redis
}

// NewChecker 新建 redis 检查器
func NewChecker(rc redis.RedisConf, opts ...redis.Option) *Checker {
	rds, err := redis.NewRedis(rc, opts...)
	if err != nil {
		panic(err)
	}

	return &Checker{rds: rds}
}

// NewCheckerWithRedis 通过已有 redis 客户端新建 redis 检查器
func NewCheckerWithRedis(rds *redis.Redis) *Checker {
	if rds == nil {
		panic(errors.New("redischecker: nil redis is invalid"))
	}

	return &Checker{rds: rds}
}

// Check 检查 redis 健康情况
func (c *Checker) Check(ctx context.Context) health.Health {
	h := health.NewHealth()

	if c.rds.PingCtx(ctx) {
		h.Up()
	} else {
		h.Down()
	}

	return h
}
