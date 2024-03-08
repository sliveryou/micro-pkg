package lock

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

/*
etcd 特性：
  etcd 可以为存储的键值对设置租约，当租约到期，键值对将失效删除。
  同时也支持续租，客户端可以在租约到期之前续约，当一个客户端持有锁期间，其它客户端只能等待，
  为避免等待期间租约失效，客户端需创建一个定时任务 KeepAlive 作为心跳不断进行续约，以避免处理还未完成而锁已经过期失效。
  在这里，concurrency.NewSession 已经帮我们完成了上面的工作。
  如果客户端在持有锁期间崩溃，心跳停止，key 将因租约到期而被删除，从而释放锁，避免死锁。

etcd 加锁解锁过程：
  1. 组装需要持有的锁名称和 LeaseID 为真正写入 etcd 的 key；
  2. 执行 put 操作，将创建的 key 绑定租约写入 etcd，客户端需记录 Revision 以便下一步判断自己是否获得锁；
  3. 通过前缀查询键值对列表，如果自己的 Revision 为当前列表中最小的则认为获得锁；否则监听列表中前一个 Revision 比自己小的 key 的删除事件，一旦监听到 pre-key 则自己获得锁；
  4. 完成业务流程后，删除对应的 key 释放锁。
*/

// Config 分布式锁相关配置
type Config struct {
	Prefix    string   `json:",optional"` // 锁前缀，如 /xlock/
	Endpoints []string `json:",optional"` // 节点列表
	Username  string   `json:",optional"` // 用户名
	Password  string   `json:",optional"` // 密码
}

// Option 可选配置
type Option func(l *Locker)

// WithClient 使用指定 etcd 客户端
func WithClient(client *clientv3.Client) Option {
	return func(l *Locker) {
		if client != nil {
			l.client = client
		}
	}
}

// Locker 分布式锁客户端
type Locker struct {
	prefix string
	client *clientv3.Client
}

// NewLocker 新建分布式锁客户端
func NewLocker(c Config, opts ...Option) (*Locker, error) {
	l := &Locker{
		prefix: strings.TrimRight(c.Prefix, "/") + "/",
	}
	for _, opt := range opts {
		opt(l)
	}

	if l.client == nil {
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   c.Endpoints,
			Username:    c.Username,
			Password:    c.Password,
			DialTimeout: time.Second * 5,
		})
		if err != nil {
			return nil, errors.WithMessage(err, "xlock: new etcd client err")
		}

		l.client = cli
	}

	return l, nil
}

// MustNewLocker 新建分布式锁客户端
func MustNewLocker(c Config, opts ...Option) *Locker {
	l, err := NewLocker(c, opts...)
	if err != nil {
		return nil
	}

	return l
}

// Lock 分布式锁
type Lock struct {
	lockKey string
	session *concurrency.Session
	mutex   *concurrency.Mutex
}

// NewLock 新建分布式锁
//
// key 为将要上锁的键，ttl 为键的租约到期时间，默认为 10s
func (l *Locker) NewLock(key string, ttl ...int) (*Lock, error) {
	t := 10
	if len(ttl) > 0 {
		t = ttl[0]
	}

	// session 可以创建租约和自动续约
	session, err := concurrency.NewSession(l.client, concurrency.WithTTL(t))
	if err != nil {
		return nil, errors.WithMessage(err, "new concurrency session err")
	}

	lockKey := l.prefix + strings.Trim(key, "/")
	mutex := concurrency.NewMutex(session, lockKey)

	return &Lock{lockKey: lockKey, session: session, mutex: mutex}, nil
}

// TryLock 尝试上锁，若不能获取锁会立刻返回
//
// ctx 最好带有 timeout
func (l *Lock) TryLock(ctx context.Context) error {
	err := l.mutex.TryLock(ctx)
	if err != nil {
		l.Close()
		return errors.WithMessage(err, "mutex try lock err")
	}

	return nil
}

// Lock 上锁，若不能获取锁会阻塞并等待获取锁
//
// ctx 最好带有 timeout
func (l *Lock) Lock(ctx context.Context) error {
	err := l.mutex.Lock(ctx)
	if err != nil {
		l.Close()
		return errors.WithMessage(err, "mutex lock err")
	}

	return nil
}

// Unlock 解锁
//
// ctx 最好带有 timeout
func (l *Lock) Unlock(ctx context.Context) error {
	defer l.Close()

	err := l.mutex.Unlock(ctx)
	if err != nil {
		return errors.WithMessage(err, "mutex unlock err")
	}

	return nil
}

// Close 关闭锁
//
// 注意：不关闭会导致 session 内存泄漏
//
// 在以下几种情况会自动调用 Close 方法：
//  1. TryLock 发生失败时
//  2. Lock 发生失败时
//  3. 执行 Unlock 时
//
// 其余情况需要自行显示调用 Close 方法
func (l *Lock) Close() {
	threading.GoSafe(func() {
		if err := l.session.Close(); err != nil {
			logx.Errorf("xlock: concurrency session close err: %v", err)
		}
	})
}
