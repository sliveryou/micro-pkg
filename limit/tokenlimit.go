package limit

import (
	"errors"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/sliveryou/go-tool/v2/timex"

	"github.com/sliveryou/micro-pkg/xkv"
)

const (
	// tokenScript 令牌桶限流 lua 脚本
	tokenScript = `local rate = tonumber(ARGV[1]);
local capacity = tonumber(ARGV[2]);
local now = tonumber(ARGV[3]);
local requested = tonumber(ARGV[4]);
local fill_time = capacity / rate;
local ttl = math.floor(fill_time * 2);
local last_info = redis.pcall('HMGET', KEYS[1], 'last_tokens', 'last_refreshed');
local last_tokens = tonumber(last_info[1]);
if last_tokens == nil then
    last_tokens = capacity;
end
local last_refreshed = tonumber(last_info[2]);
if last_refreshed == nil then
    last_refreshed = 0;
end
local delta = math.max(0, now - last_refreshed);
local filled_tokens = math.min(capacity, last_tokens + (delta * rate));
local allowed = filled_tokens >= requested;
local new_tokens = filled_tokens;
if allowed then
    new_tokens = filled_tokens - requested;
end
redis.call('HMSET', KEYS[1], 'last_tokens', new_tokens, 'last_refreshed', now);
redis.call('EXPIRE', KEYS[1], ttl);
return allowed;`
)

// tokenOption 令牌桶限流器配置
type tokenOption struct {
	rate     int    // 速率，即每秒生成的令牌数量
	capacity int    // 容量，即令牌桶所能容纳的令牌的最大数量
	key      string // 键
}

// clone 克隆令牌桶限流器配置
func (o tokenOption) clone() *tokenOption {
	return &tokenOption{
		rate:     o.rate,
		capacity: o.capacity,
		key:      o.key,
	}
}

// TokenOption 令牌桶限流器可选配置
type TokenOption func(op *tokenOption)

// WithRate 指定令牌桶限流器速率
func WithRate(rate int) TokenOption {
	return func(op *tokenOption) {
		op.rate = rate
	}
}

// WithCapacity 指定令牌桶限流器容量
func WithCapacity(capacity int) TokenOption {
	return func(op *tokenOption) {
		op.capacity = capacity
	}
}

// WithKey 指定令牌桶限流器键
func WithKey(key string) TokenOption {
	return func(op *tokenOption) {
		op.key = key
	}
}

// TokenLimit 令牌桶限流器结构详情
type TokenLimit struct {
	option tokenOption
	store  *xkv.Store
}

// NewTokenLimit 新建令牌桶限流器
func NewTokenLimit(rate, capacity int, key string, store *xkv.Store) (*TokenLimit, error) {
	if store == nil || rate <= 0 || capacity <= 0 {
		return nil, errors.New("limit: illegal token limit config")
	}

	limiter := &TokenLimit{
		option: tokenOption{
			rate:     rate,
			capacity: capacity,
			key:      key,
		},
		store: store,
	}

	return limiter, nil
}

// MustNewTokenLimit 新建令牌桶限流器
func MustNewTokenLimit(rate, capacity int, key string, store *xkv.Store) *TokenLimit {
	tl, err := NewTokenLimit(rate, capacity, key, store)
	if err != nil {
		panic(err)
	}

	return tl
}

// Allow 访问令牌桶限流器拿取令牌，并判断是否允许拿取
func (tl *TokenLimit) Allow(opts ...TokenOption) (bool, error) {
	return tl.AllowN(timex.Now(), 1, opts...)
}

// AllowN 访问令牌桶限流器拿取 n 个令牌，并判断是否允许拿取
func (tl *TokenLimit) AllowN(now time.Time, n int, opts ...TokenOption) (bool, error) {
	op := tl.option.clone()
	for _, opt := range opts {
		opt(op)
	}

	// 类型转换参考：http://doc.redisfans.com/script/eval.html#lua-redis
	resp, err := tl.store.Eval(tokenScript, op.key,
		op.rate, op.capacity, now.Unix(), n,
	)
	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	code, ok := resp.(int64)
	if !ok {
		return false, ErrUnexpectedType
	}

	return code == 1, nil
}
