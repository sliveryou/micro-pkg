package xkv

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/kv"

	"github.com/sliveryou/go-tool/v2/convert"
)

const (
	// getAndDelScript 获取并删除 key 所关联的值 lua 脚本
	getAndDelScript = `local current = redis.call('GET', KEYS[1]);
if (current) then
    redis.call('DEL', KEYS[1]);
end
return current;`
)

// Store 键值存取器结构详情
type Store struct {
	c kv.KvConf
	kv.Store
}

// NewStore 新建键值存取器
func NewStore(c kv.KvConf) *Store {
	return &Store{c: c, Store: kv.NewStore(c)}
}

// GetInt 返回给定 key 所关联的 int 值
func (s *Store) GetInt(key string) (int, error) {
	return s.GetIntCtx(context.Background(), key)
}

// GetIntCtx 返回给定 key 所关联的 int 值
func (s *Store) GetIntCtx(ctx context.Context, key string) (int, error) {
	value, err := s.GetCtx(ctx, key)
	if err != nil {
		return 0, errors.Wrap(err, "get err")
	}

	return convert.ToInt(value), nil
}

// SetInt 将 int value 关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) SetInt(key string, value int, seconds ...int) error {
	return s.SetIntCtx(context.Background(), key, value, seconds...)
}

// SetIntCtx 将 int value 关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) SetIntCtx(ctx context.Context, key string, value int, seconds ...int) error {
	return s.SetStringCtx(ctx, key, convert.ToString(value), seconds...)
}

// GetInt64 返回给定 key 所关联的 int64 值
func (s *Store) GetInt64(key string) (int64, error) {
	return s.GetInt64Ctx(context.Background(), key)
}

// GetInt64Ctx 返回给定 key 所关联的 int64 值
func (s *Store) GetInt64Ctx(ctx context.Context, key string) (int64, error) {
	value, err := s.GetCtx(ctx, key)
	if err != nil {
		return 0, errors.Wrap(err, "get err")
	}

	return convert.ToInt64(value), nil
}

// SetInt64 将 int64 value 关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) SetInt64(key string, value int64, seconds ...int) error {
	return s.SetInt64Ctx(context.Background(), key, value, seconds...)
}

// SetInt64Ctx 将 int64 value 关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) SetInt64Ctx(ctx context.Context, key string, value int64, seconds ...int) error {
	return s.SetStringCtx(ctx, key, convert.ToString(value), seconds...)
}

// GetBytes 返回给定 key 所关联的 []byte 值
func (s *Store) GetBytes(key string) ([]byte, error) {
	return s.GetBytesCtx(context.Background(), key)
}

// GetBytesCtx 返回给定 key 所关联的 []byte 值
func (s *Store) GetBytesCtx(ctx context.Context, key string) ([]byte, error) {
	value, err := s.GetCtx(ctx, key)
	if err != nil {
		return nil, errors.Wrap(err, "get err")
	}

	return []byte(value), nil
}

// GetDel 返回并删除给定 key 所关联的 string 值
func (s *Store) GetDel(key string) (string, error) {
	return s.GetDelCtx(context.Background(), key)
}

// GetDelCtx 返回并删除给定 key 所关联的 string 值
func (s *Store) GetDelCtx(ctx context.Context, key string) (string, error) {
	resp, err := s.EvalCtx(ctx, getAndDelScript, key)
	if err != nil {
		return "", errors.Wrap(err, "eval script err")
	}

	return convert.ToString(resp), nil
}

// SetString 将 string value 关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) SetString(key, value string, seconds ...int) error {
	return s.SetStringCtx(context.Background(), key, value, seconds...)
}

// SetStringCtx 将 string value 关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) SetStringCtx(ctx context.Context, key, value string, seconds ...int) error {
	if len(seconds) > 0 {
		return errors.Wrapf(s.SetexCtx(ctx, key, value, seconds[0]),
			"setex by seconds: %d err", seconds[0])
	}

	return errors.Wrap(s.SetCtx(ctx, key, value), "set err")
}

// Read 将给定 key 所关联的值反序列化到 obj 对象
//
// 返回 false 时代表给定 key 不存在
func (s *Store) Read(key string, obj any) (bool, error) {
	return s.ReadCtx(context.Background(), key, obj)
}

// ReadCtx 将给定 key 所关联的值反序列化到 obj 对象
//
// 返回 false 时代表给定 key 不存在
func (s *Store) ReadCtx(ctx context.Context, key string, obj any) (bool, error) {
	if !isValid(obj) {
		return false, errors.New("obj is invalid")
	}

	value, err := s.GetBytesCtx(ctx, key)
	if err != nil {
		return false, errors.Wrap(err, "get bytes err")
	}
	if len(value) == 0 {
		return false, nil
	}

	err = json.Unmarshal(value, obj)
	if err != nil {
		return false, errors.Wrap(err, "json unmarshal value to obj err")
	}

	return true, nil
}

// Write 将对象 obj 序列化后关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) Write(key string, obj any, seconds ...int) error {
	return s.WriteCtx(context.Background(), key, obj, seconds...)
}

// WriteCtx 将对象 obj 序列化后关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) WriteCtx(ctx context.Context, key string, obj any, seconds ...int) error {
	value, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "json marshal obj err")
	}

	return s.SetStringCtx(ctx, key, string(value), seconds...)
}

// ReadOrGet 将给定 key 所关联的值反序列化到 obj 对象
//
//	若给定 key 不存在则调用数据获取函数，调用成功时赋值至 obj 对象，并将其序列化后关联到给定 key
//	seconds 为 key 的过期时间（秒）
func (s *Store) ReadOrGet(key string, obj any, gf func() (any, error), seconds ...int) error {
	f := func(context.Context) (any, error) { return gf() }
	return s.ReadOrGetCtx(context.Background(), key, obj, f, seconds...)
}

// ReadOrGetCtx 将给定 key 所关联的值反序列化到 obj 对象
//
//	若给定 key 不存在则调用数据获取函数，调用成功时赋值至 obj 对象，并将其序列化后关联到给定 key
//	seconds 为 key 的过期时间（秒）
func (s *Store) ReadOrGetCtx(ctx context.Context, key string, obj any, gf func(ctx context.Context) (any, error), seconds ...int) error {
	isExist, err := s.ReadCtx(ctx, key, obj)
	if err != nil {
		return errors.Wrap(err, "read obj by err")
	}

	if !isExist {
		data, err := gf(ctx)
		if err != nil {
			return err
		}

		if !isValid(data) {
			return errors.New("get data is invalid")
		}

		ov, dv := reflect.ValueOf(obj).Elem(), reflect.ValueOf(data).Elem()
		if ov.Type() != dv.Type() {
			return errors.New("obj type and get data type are not equal")
		}
		ov.Set(dv)

		_ = s.WriteCtx(ctx, key, data, seconds...)
	}

	return nil
}

// isValid 判断对象是否合法
func isValid(obj any) bool {
	if obj == nil {
		return false
	}

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return false
	}

	return val.Elem().CanAddr()
}
