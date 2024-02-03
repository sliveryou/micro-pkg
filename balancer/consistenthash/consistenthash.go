package consistenthash

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"

	"github.com/sliveryou/go-tool/v2/randx"

	"github.com/sliveryou/micro-pkg/xhash/ketama"
)

const (
	// Name 一致性 hash 平衡器的注册名称
	Name = "consistent_hash"
	// DefaultKey 默认的用于计算一致性 hash 的 key
	DefaultKey = ContextKey("X-Consistent-Hash")
)

// ContextKey 上下文 key 类型
type ContextKey string

// String 实现序列化字符串方法
func (c ContextKey) String() string {
	return "consistent hash context key: " + string(c)
}

// RegisterBuilder 注册一致性 hash 构建器
func RegisterBuilder(chKey ...ContextKey) {
	balancer.Register(newBuilder(chKey...))
}

// newBuilder 新建一致性 hash 构建器
func newBuilder(chKey ...ContextKey) balancer.Builder {
	ck := DefaultKey
	if len(chKey) > 0 {
		ck = chKey[0]
	}

	return base.NewBalancerBuilder(Name,
		&chPickerBuilder{chKey: ck},
		base.Config{HealthCheck: true},
	)
}

// chPickerBuilder 一致性 hash 构建器
type chPickerBuilder struct {
	chKey ContextKey
}

// Build 构建一致性 hash 选取器
func (b *chPickerBuilder) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
	if len(buildInfo.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	// 构建 chPicker
	picker := &chPicker{
		subConns: make(map[string]balancer.SubConn),
		ch:       ketama.New(), // 基于 Ketama 算法的一致性 hash 负载均衡器
		chKey:    b.chKey,      // 用于计算一致性 hash 的 key
	}

	for sc, conInfo := range buildInfo.ReadySCs {
		node := conInfo.Address.Addr
		picker.ch.Add(node)
		picker.subConns[node] = sc
	}

	return picker
}

// chPicker 一致性 hash 选取器
type chPicker struct {
	subConns map[string]balancer.SubConn
	ch       *ketama.Ketama
	chKey    ContextKey
}

// Pick 获取一致性 hash 选取结果
func (p *chPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var ret balancer.PickResult

	key, _ := info.Ctx.Value(p.chKey).(string)
	if key == "" {
		key = randx.NewString(20)
	}

	node, ok := p.ch.Get(key)
	if ok {
		ret.SubConn = p.subConns[node]
	}

	return ret, nil
}
