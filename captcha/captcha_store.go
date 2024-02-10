package captcha

import (
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
	return &Store{kvStore: kvStore, keyPrefix: keyPrefix, expiration: expiration}
}

// Set 存储验证码信息
func (s *Store) Set(id, value string) error {
	key := s.keyPrefix + id
	err := s.kvStore.SetString(key, value, s.expiration)
	if err != nil {
		logx.Errorf("captcha: Store.Set err: %v", err)
		return err
	}

	return nil
}

// Get 获取验证码信息
func (s *Store) Get(id string, clear bool) string {
	key := s.keyPrefix + id
	value, err := s.kvStore.Get(key)
	if err != nil {
		logx.Errorf("captcha: Store.Get err: %v", err)
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
		logx.Errorf("captcha: Store.Verify err: %v", err)
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
