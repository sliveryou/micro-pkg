package disabler

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/pkg/errors"
)

// Adapter 决策规则适配器
type Adapter struct {
	Routes []string
}

// NewAdapter 新建决策规则适配器
func NewAdapter(routes []string) *Adapter {
	return &Adapter{Routes: routes}
}

// LoadPolicy 加载决策规则
func (a *Adapter) LoadPolicy(m model.Model) error {
	// route 常见形式：
	//   GET:/api/user/{id}
	//   /api/user
	//   /user.User/GetUser
	for _, route := range a.Routes {
		if route == "" {
			continue
		}

		var obj, act string
		parts := strings.SplitN(route, ":", 2)
		if length := len(parts); length == 1 {
			obj, act = parts[0], "*"
		} else if length == 2 {
			obj, act = parts[1], parts[0]
		}
		line := fmt.Sprintf("p, %s, %s, allow", obj, act)

		if err := persist.LoadPolicyLine(line, m); err != nil {
			return errors.WithMessagef(err, "load policy line: %s err", line)
		}
	}

	return nil
}

// SavePolicy 保存决策规则
func (a *Adapter) SavePolicy(m model.Model) error {
	return nil
}

// AddPolicy 添加决策规则
func (a *Adapter) AddPolicy(sec, ptype string, rule []string) error {
	return nil
}

// RemovePolicy 移除决策规则
func (a *Adapter) RemovePolicy(sec, ptype string, rule []string) error {
	return nil
}

// RemoveFilteredPolicy 移除筛选后的决策规则
func (a *Adapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}
