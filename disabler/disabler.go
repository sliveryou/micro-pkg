package disabler

import (
	casbin "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/pkg/errors"
)

const (
	// DefaultModelText 默认 casbin 模型文本
	DefaultModelText = `
[request_definition]
r = obj, act

[policy_definition]
p = obj, act, eft

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = keyMatch3(r.obj, p.obj) && (r.act == p.act || p.act == "*")
`
)

// Config 功能禁用器相关配置
type Config struct {
	DisabledAPIs []string `json:",optional"` // API 禁用列表
	DisabledRPCs []string `json:",optional"` // RPC 禁用列表
}

// FuncDisabler 功能禁用器
type FuncDisabler struct {
	c           Config
	apiEnforcer *casbin.Enforcer
	rpcEnforcer *casbin.Enforcer
}

// NewFuncDisabler 新建功能禁用器
func NewFuncDisabler(c Config) (*FuncDisabler, error) {
	apiEnforcer, err := newEnforcer(c.DisabledAPIs)
	if err != nil {
		return nil, errors.New("disabler: new api enforcer err")
	}

	rpcEnforcer, err := newEnforcer(c.DisabledRPCs)
	if err != nil {
		return nil, errors.New("disabler: new rpc enforcer err")
	}

	return &FuncDisabler{
		c:           c,
		apiEnforcer: apiEnforcer,
		rpcEnforcer: rpcEnforcer,
	}, nil
}

// MustNewFuncDisabler 新建功能禁用器
func MustNewFuncDisabler(c Config) *FuncDisabler {
	fd, err := NewFuncDisabler(c)
	if err != nil {
		panic(err)
	}

	return fd
}

// AllowAPI 是否允许放行该 API 请求
func (fd *FuncDisabler) AllowAPI(method, api string) bool {
	hit, _ := fd.apiEnforcer.Enforce(api, method)
	return !hit
}

// AllowRPC 是否允许放行该 RPC 请求
func (fd *FuncDisabler) AllowRPC(rpc string) bool {
	hit, _ := fd.rpcEnforcer.Enforce(rpc, "*")
	return !hit
}

// newEnforcer 新建决策执行器
func newEnforcer(routes []string) (*casbin.Enforcer, error) {
	a := NewAdapter(routes)
	m, err := model.NewModelFromString(DefaultModelText)
	if err != nil {
		return nil, errors.WithMessage(err, "new model from string err")
	}

	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, errors.WithMessage(err, "new enforcer err")
	}

	return e, nil
}
