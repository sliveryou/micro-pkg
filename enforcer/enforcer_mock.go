package enforcer

import (
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// MockAdapter 模拟决策规则适配器
type MockAdapter struct{}

// LoadPolicy 加载决策规则
func (a *MockAdapter) LoadPolicy(m model.Model) error {
	for _, line := range getMockPolicies() {
		s := strings.Split(line, ",")
		if len(s) >= 5 {
			s = s[:5]
		}
		p := strings.Join(s, ",")

		if err := persist.LoadPolicyLine(p, m); err != nil {
			return err
		}
	}

	return nil
}

// SavePolicy 保存决策规则
func (a *MockAdapter) SavePolicy(m model.Model) error {
	return nil
}

// AddPolicy 添加决策规则
func (a *MockAdapter) AddPolicy(sec, ptype string, rule []string) error {
	return nil
}

// RemovePolicy 移除决策规则
func (a *MockAdapter) RemovePolicy(sec, ptype string, rule []string) error {
	return nil
}

// RemoveFilteredPolicy 移除筛选后的决策规则
func (a *MockAdapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

// getMockPolicies 获取模拟决策规则
func getMockPolicies() []string {
	return []string{
		"p, ADMIN, /api/department, GET, allow, 查询部门分页",
		"p, ADMIN, /api/department, POST, allow, 创建部门",
		"p, ADMIN, /api/department/{id}, GET, allow, 查询部门",
		"p, ADMIN, /api/department/{id}, PUT, allow, 更新部门",
		"p, ADMIN, /api/department/{id}, DELETE, allow, 删除部门",
		"p, ADMIN, /api/job, GET, allow, 查询岗位分页",
		"p, ADMIN, /api/job, POST, allow, 创建岗位",
		"p, ADMIN, /api/job/{id}, GET, allow, 查询岗位",
		"p, ADMIN, /api/job/{id}, PUT, allow, 更新岗位",
		"p, ADMIN, /api/job/{id}, DELETE, allow, 删除岗位",
		"p, ADMIN, /api/personnel, GET, allow, 查询人员分页",
		"p, ADMIN, /api/personnel, POST, allow, 创建人员",
		"p, ADMIN, /api/personnel/{id}, GET, allow, 查询人员",
		"p, ADMIN, /api/personnel/{id}, PUT, allow, 更新人员",
		"p, ADMIN, /api/personnel/{id}, DELETE, allow, 删除人员",
	}
}

// MockWatcher 模拟决策规则观察器
type MockWatcher struct {
	callback func(string)
}

// SetUpdateCallback 设置更新回调函数
func (w *MockWatcher) SetUpdateCallback(callback func(string)) error {
	w.callback = callback

	return nil
}

// Update 调用更新回调函数
func (w *MockWatcher) Update() error {
	if w.callback != nil {
		w.callback("")
	}

	return nil
}

// Close 关闭模拟决策规则观察器
func (w *MockWatcher) Close() {}
