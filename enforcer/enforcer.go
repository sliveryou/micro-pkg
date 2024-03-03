package enforcer

import (
	"strings"
	"sync/atomic"
	"time"

	casbin "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	cutil "github.com/casbin/casbin/v2/util"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

const (
	// DefaultModelText 默认 casbin 模型文本
	DefaultModelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && URIMatch(r.obj, p.obj) && r.act == p.act
`
)

// Config 决策规则执行器配置
type Config struct {
	ModelText     string        `json:",optional"`      // casbin 模型文本，为空则使用 DefaultModelText
	RetryDuration time.Duration `json:",default=500ms"` // 加载决策规则重试间隔
	RetryMaxTimes int           `json:",default=5"`     // 加载决策规则重试最大次数
}

// Enforcer 决策规则执行器
type Enforcer struct {
	*casbin.SyncedEnforcer
	c           Config
	a           persist.Adapter
	w           persist.Watcher
	loadRunning int32
}

// NewEnforcer 新建决策规则执行器
func NewEnforcer(c Config, a persist.Adapter, w persist.Watcher) (*Enforcer, error) {
	if a == nil || w == nil {
		return nil, errors.New("enforcer: illegal enforcer config")
	}
	if err := c.fillDefault(); err != nil {
		return nil, errors.WithMessage(err, "enforcer: fill default config err")
	}

	e := &Enforcer{c: c, a: a, w: w, loadRunning: 0}
	if err := e.init(); err != nil {
		return nil, errors.WithMessage(err, "enforcer: init enforcer err")
	}

	return e, nil
}

// MustNewEnforcer 新建决策规则执行器
func MustNewEnforcer(c Config, a persist.Adapter, w persist.Watcher) *Enforcer {
	e, err := NewEnforcer(c, a, w)
	if err != nil {
		panic(err)
	}

	return e
}

// init 初始化决策规则执行器
func (e *Enforcer) init() error {
	m, err := model.NewModelFromString(e.c.ModelText)
	if err != nil {
		return errors.WithMessage(err, "model.NewModelFromString err")
	}

	se, err := casbin.NewSyncedEnforcer(m, e.a)
	if err != nil {
		return errors.WithMessage(err, "casbin.NewSyncedEnforcer err")
	}

	if err := se.SetWatcher(e.w); err != nil {
		return errors.WithMessage(err, "se.SetWatcher err")
	}

	if err := se.LoadPolicy(); err != nil {
		return errors.WithMessage(err, "se.LoadPolicy err")
	}

	se.AddFunction("URIMatch", URIMatchWrapper)
	e.SyncedEnforcer = se

	_ = e.w.SetUpdateCallback(func(string) {
		e.Reload(e.c.RetryDuration, e.c.RetryMaxTimes)
	})

	return nil
}

// Update 更新决策规则信息
func (e *Enforcer) Update() {
	err := e.w.Update()
	if err != nil {
		logx.Errorf("enforcer: watcher update err: %v", err)
	}
}

// Reload 重新加载决策规则
func (e *Enforcer) Reload(retryDuration time.Duration, retryMaxTimes int) {
	if atomic.LoadInt32(&e.loadRunning) != 0 {
		return
	}
	atomic.StoreInt32(&e.loadRunning, int32(1))

	var err error
	ticker := time.NewTicker(retryDuration)

	threading.GoSafe(func() {
		defer func() {
			ticker.Stop()
			atomic.StoreInt32(&e.loadRunning, int32(0))
			if err != nil {
				logx.Errorf("enforcer: reload polocy err: %v", err)
			}
		}()

		retryTimes := 0
		max := make(chan int)

		for {
			select {
			case <-ticker.C:
				retryTimes++
				if err = e.LoadPolicy(); err == nil {
					logx.Infof("enforcer: reload polocy successfully, retryTimes: %d", retryTimes)
					return
				}

				if retryTimes >= retryMaxTimes {
					max <- retryTimes
				}
			case <-max:
				return
			}
		}
	})
}

// URIMatch URI 决策规则函数
func URIMatch(key1, key2 string) bool {
	key1 = strings.Split(key1, "?")[0]

	return cutil.KeyMatch3(key1, key2)
}

// URIMatchWrapper URI 决策规则函数装饰器
func URIMatchWrapper(args ...any) (any, error) {
	var key1, key2 string
	if len(args) >= 2 {
		key1, _ = args[0].(string)
		key2, _ = args[1].(string)
	}

	return URIMatch(key1, key2), nil
}

// fillDefault 填充默认值
func (c *Config) fillDefault() error {
	fill := &Config{}
	if err := conf.FillDefault(fill); err != nil {
		return err
	}

	if c.ModelText == "" {
		c.ModelText = DefaultModelText
	}

	return mergo.Merge(c, fill)
}
