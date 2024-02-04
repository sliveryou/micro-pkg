package compositechecker

import (
	"context"
	"sync"

	"github.com/sliveryou/micro-pkg/health"
)

var _ health.Checker = (*Checker)(nil)

// checkerItem 应用检查器结构封装
type checkerItem struct {
	name    string
	checker health.Checker
}

// Checker 复合应用检查器
type Checker struct {
	mu         sync.RWMutex
	checkers   []*checkerItem
	checkerMap map[string]*checkerItem
	info       map[string]any
}

// NewChecker 新建复合应用检查器
func NewChecker() *Checker {
	return &Checker{
		checkerMap: make(map[string]*checkerItem),
		info:       make(map[string]any),
	}
}

// AddInfo 添加健康键值信息
func (c *Checker) AddInfo(key string, value any) *Checker {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.info[key] = value

	return c
}

// AddChecker 添加一个应用检查器（切记 name 不能重复）
func (c *Checker) AddChecker(name string, checker health.Checker) *Checker {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.checkerMap[name]; !ok {
		ci := &checkerItem{name: name, checker: checker}
		c.checkers = append(c.checkers, ci)
		c.checkerMap[name] = ci
	}

	return c
}

// Check 检查所有应用健康情况，当某个应用健康情况不为 Up，则复合应用健康情况为 Down
func (c *Checker) Check(ctx context.Context) health.Health {
	c.mu.RLock()
	defer c.mu.RUnlock()

	h := health.NewHealth()
	h.Up()

	type state struct {
		h    health.Health
		name string
	}
	ch := make(chan state, len(c.checkers))
	var wg sync.WaitGroup
	for _, item := range c.checkers {
		wg.Add(1)
		item := item
		go func() {
			ch <- state{h: item.checker.Check(ctx), name: item.name}
			wg.Done()
		}()
	}
	wg.Wait()
	close(ch)

	for s := range ch {
		if !s.h.IsUp() && !h.IsDown() {
			h.Down()
		}
		h.AddInfo(s.name, s.h)
	}

	// 额外信息
	for key, value := range c.info {
		h.AddInfo(key, value)
	}

	return h
}

// CheckByName 检查指定应用健康情况，当找不到该应用时，应用健康情况为 Unknown
func (c *Checker) CheckByName(ctx context.Context, name string) health.Health {
	c.mu.RLock()
	defer c.mu.RUnlock()

	h := health.NewHealth()
	ci, ok := c.checkerMap[name]
	if !ok {
		h.Unknown().AddInfo("error", "unknown service name")
		return h
	}

	return ci.checker.Check(ctx)
}
