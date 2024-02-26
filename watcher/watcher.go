package watcher

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/threading"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config etcd 观察器配置
type Config struct {
	Key       string   // 需要观察的键
	Endpoints []string `json:",optional"` // 节点列表
	Username  string   `json:",optional"` // 用户名
	Password  string   `json:",optional"` // 密码
}

// Option 可选配置
type Option func(w *Watcher)

// WithClient 使用指定 etcd 客户端
func WithClient(client *clientv3.Client) Option {
	return func(w *Watcher) {
		if client != nil {
			w.client = client
		}
	}
}

// Watcher etcd 观察器
type Watcher struct {
	c           Config
	lock        sync.RWMutex
	client      *clientv3.Client
	cancel      context.CancelFunc
	lastSentRev int64
	callback    func(string)
}

// NewWatcher 新建 etcd 观察器
func NewWatcher(c Config, opts ...Option) (*Watcher, error) {
	if c.Key == "" {
		return nil, errors.New("watcher: illegal watcher config")
	}

	w := &Watcher{c: c}
	for _, opt := range opts {
		opt(w)
	}

	if w.client == nil {
		if err := w.createClient(); err != nil {
			return nil, errors.WithMessage(err, "watcher: create client err")
		}
	}

	w.startWatch()

	return w, nil
}

// MustNewWatcher 新建 etcd 观察器
func MustNewWatcher(c Config, opts ...Option) *Watcher {
	w, err := NewWatcher(c, opts...)
	if err != nil {
		panic(err)
	}

	return w
}

// SetUpdateCallback 设置 etcd 更新回调函数
func (w *Watcher) SetUpdateCallback(callback func(string)) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.callback = callback

	return nil
}

// Update 触发 etcd 更新事件
func (w *Watcher) Update() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := w.client.Put(ctx, w.c.Key, "")
	if err == nil {
		w.lastSentRev = resp.Header.GetRevision()
	}

	return err
}

// Close 关闭 etcd 观察器
func (w *Watcher) Close() {
	w.cancel()
}

// createClient 创建 etcd 客户端
func (w *Watcher) createClient() error {
	cfg := clientv3.Config{
		Endpoints:            w.c.Endpoints,
		Username:             w.c.Username,
		Password:             w.c.Password,
		AutoSyncInterval:     60 * time.Second,
		DialTimeout:          10 * time.Second,
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 10 * time.Second,
	}

	c, err := clientv3.New(cfg)
	if err != nil {
		return errors.WithMessage(err, "new etcd client err")
	}

	w.client = c

	return nil
}

// startWatch 监听 etcd 更新事件
func (w *Watcher) startWatch() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	threading.GoSafe(func() {
		defer w.cancel()

		wch := w.client.Watch(ctx, w.c.Key)
		for wr := range wch {
			for _, ev := range wr.Events {
				// 监听创建和更新事件
				if ev.IsCreate() || ev.IsModify() {
					w.lock.RLock()
					if rev := ev.Kv.ModRevision; w.callback != nil {
						w.callback(strconv.FormatInt(rev, 10))
					}
					w.lock.RUnlock()
				}
			}
		}
	})
}
