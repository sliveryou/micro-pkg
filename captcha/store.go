package captcha

import (
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/sliveryou/micro-pkg/xkv"
)

// Store 验证码存储器结构详情
type Store struct {
	kvStore    *xkv.Store // 键值存取器
	keyPrefix  string     // 存储验证码的 key 前缀
	expiration int        // 存储验证码的 value 过期时间，单位（秒）
}

// NewStore 新建验证码存储器
func NewStore(kvStore *xkv.Store, keyPrefix string, expiration int) *Store {
	if kvStore == nil {
		panic(errors.New("captcha: illegal store config"))
	}

	return &Store{kvStore: kvStore, keyPrefix: keyPrefix, expiration: expiration}
}

// Set 存储验证码信息
func (s *Store) Set(id, value string) error {
	key := s.keyPrefix + id
	err := s.kvStore.SetString(key, value, s.expiration)
	if err != nil {
		logx.Errorf("captcha: store set err: %v, key: %s", err, key)
		return err
	}

	return nil
}

// Get 获取验证码信息
func (s *Store) Get(id string, clear bool) string {
	key := s.keyPrefix + id
	value, err := s.kvStore.Get(key)
	if err != nil {
		logx.Errorf("captcha: store get err: %v, key: %s", err, key)
		return ""
	}

	if clear {
		s.kvStore.Del(key)
	}

	return value
}

// Verify 校验验证码信息
func (s *Store) Verify(id, answer string, clear bool) bool {
	key := s.keyPrefix + id
	value, err := s.kvStore.Get(key)
	if err != nil {
		logx.Errorf("captcha: store verify err: %v, key: %s", err, key)
		return false
	}

	if answer != value {
		return false
	}

	if clear {
		s.kvStore.Del(key)
	}

	return true
}
