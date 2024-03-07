package limit

import (
	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/xkv"
)

const (
	// takeScript take 时间段限流 lua 脚本
	takeScript = `local limit = tonumber(ARGV[1]);
local window = tonumber(ARGV[2]);
local current = redis.call("INCRBY", KEYS[1], 1);
if current == 1 then
    redis.call("EXPIRE", KEYS[1], window);
    return 1;
elseif current < limit then
    return 1;
elseif current == limit then
    return 2;
else
    return 3;
end`

	// getScript get 时间段限流 lua 脚本
	getScript = `local limit = tonumber(ARGV[1]);
local window = tonumber(ARGV[2]);
local current = redis.call("INCRBY", KEYS[1], 1);
if current == 1 then
    redis.call("EXPIRE", KEYS[1], window);
end
return limit - current;
`
)

const (
	// Unknown 未知状态码
	Unknown = 0
	// Allowed 小于配额状态码
	Allowed = 1
	// Reached 达到配额状态码
	Reached = 2
	// Overflowed 超出配额状态码
	Overflowed = 3

	// InvalidQuota 无效配额
	InvalidQuota = -1

	internalAllowed   = 1
	internalHitQuota  = 2
	internalOverQuota = 3
)

var (
	// ErrUnknownCode 未知状态码错误
	ErrUnknownCode = errors.New("unknown code")
	// ErrUnexpectedType 意外类型错误
	ErrUnexpectedType = errors.New("unexpected type")
)

// option 限流器配置
type option struct {
	period    int    // 时间段
	quota     int    // 配额
	keyPrefix string // 键前缀
}

// clone 克隆限流器配置
func (o option) clone() *option {
	return &option{
		period:    o.period,
		quota:     o.quota,
		keyPrefix: o.keyPrefix,
	}
}

// Option 限流器可选配置
type Option func(op *option)

// WithPeriod 指定限流器时间段
func WithPeriod(period int) Option {
	return func(op *option) {
		op.period = period
	}
}

// WithQuota 指定限流器配额
func WithQuota(quota int) Option {
	return func(op *option) {
		op.quota = quota
	}
}

// WithKeyPrefix 指定限流器键前缀
func WithKeyPrefix(keyPrefix string) Option {
	return func(op *option) {
		op.keyPrefix = keyPrefix
	}
}

// PeriodLimit 时间段限流器结构详情
type PeriodLimit struct {
	option option
	store  *xkv.Store
}

// NewPeriodLimit 新建时间段限流器
func NewPeriodLimit(period, quota int, keyPrefix string, store *xkv.Store) (*PeriodLimit, error) {
	if store == nil || period <= 0 || quota <= 0 {
		return nil, errors.New("limit: illegal period limit config")
	}

	limiter := &PeriodLimit{
		option: option{
			period:    period,
			quota:     quota,
			keyPrefix: keyPrefix,
		},
		store: store,
	}

	return limiter, nil
}

// MustNewPeriodLimit 新建时间段限流器
func MustNewPeriodLimit(period, quota int, keyPrefix string, store *xkv.Store) *PeriodLimit {
	pl, err := NewPeriodLimit(period, quota, keyPrefix, store)
	if err != nil {
		panic(err)
	}

	return pl
}

// Take 访问时间段限流器拿取配额，并返回相关状态码
func (pl *PeriodLimit) Take(key string, opts ...Option) (int, error) {
	op := pl.option.clone()
	for _, opt := range opts {
		opt(op)
	}

	resp, err := pl.store.Eval(takeScript, op.keyPrefix+key,
		op.quota, op.period,
	)
	if err != nil {
		return Unknown, errors.WithMessage(err, "store eval script err")
	}

	code, ok := resp.(int64)
	if !ok {
		return Unknown, ErrUnexpectedType
	}

	switch code {
	case internalAllowed:
		return Allowed, nil
	case internalHitQuota:
		return Reached, nil
	case internalOverQuota:
		return Overflowed, nil
	default:
		return Unknown, ErrUnknownCode
	}
}

// Get 访问时间段限流器拿取配额，并返回剩余配额，返回值含义
//
//	> 0：剩余配额大于 0，Allowed
//	= 0：剩余配额等于 0，Reached
//	< 0：剩余配额小于 0，Overflowed
func (pl *PeriodLimit) Get(key string, opts ...Option) (int, error) {
	op := pl.option.clone()
	for _, opt := range opts {
		opt(op)
	}

	resp, err := pl.store.Eval(getScript, op.keyPrefix+key,
		op.quota, op.period,
	)
	if err != nil {
		return InvalidQuota, errors.WithMessage(err, "store eval script err")
	}

	remain, ok := resp.(int64)
	if !ok {
		return InvalidQuota, ErrUnexpectedType
	}

	return int(remain), nil
}

// Allow 访问时间段限流器拿取配额，并判断是否允许放行
func (pl *PeriodLimit) Allow(key string, opts ...Option) (bool, error) {
	code, err := pl.Take(key, opts...)
	if err != nil {
		return false, err
	}

	if code == Allowed || code == Reached {
		return true, nil
	}

	return false, nil
}
