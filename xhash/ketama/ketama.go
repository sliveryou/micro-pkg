package ketama

import (
	"sort"
	"strconv"
	"sync"

	"github.com/spaolacci/murmur3"
)

const (
	// DefaultReplicas 默认的虚拟节点数
	DefaultReplicas = 100
	// MaxWeight 最大节点权重
	MaxWeight = DefaultReplicas
)

// HashFunc hash 生成函数
type HashFunc func(data []byte) uint64

// DefaultHash 默认 hash 生成函数
func DefaultHash(data []byte) uint64 {
	return murmur3.Sum64(data)
}

// Ketama 基于 Ketama 算法的一致性 hash 负载均衡器
type Ketama struct {
	sync.RWMutex                   // 读写锁
	replicas     int               // 每个真实节点对应复制的虚拟节点数
	hashFunc     HashFunc          // hash 生成函数
	ring         []uint64          // 排序后的 hash 环
	nodeMap      map[uint64]string // 虚拟节点到真实节点的映射
}

// New 新建一个 Ketama 对象
func New() *Ketama {
	return NewCustom(DefaultReplicas, DefaultHash)
}

// NewCustom 新建一个自定义 Ketama 对象
func NewCustom(replicas int, hashFunc HashFunc) *Ketama {
	k := &Ketama{
		replicas: replicas,
		hashFunc: hashFunc,
		nodeMap:  make(map[uint64]string),
	}

	if k.replicas < DefaultReplicas {
		k.replicas = DefaultReplicas
	}

	if k.hashFunc == nil {
		k.hashFunc = DefaultHash
	}

	return k
}

// AddWithReplicas 添加带虚拟节点的节点
func (k *Ketama) AddWithReplicas(node string, replicas int) {
	if replicas < 1 {
		replicas = 1
	} else if replicas > k.replicas {
		replicas = k.replicas
	}

	k.Lock()
	defer k.Unlock()

	for i := 0; i < replicas; i++ {
		hash := k.hashFunc([]byte(node + strconv.Itoa(i)))

		if _, ok := k.nodeMap[hash]; !ok {
			k.ring = append(k.ring, hash)
		}

		k.nodeMap[hash] = node
	}

	sort.Slice(k.ring, func(i, j int) bool {
		return k.ring[i] < k.ring[j]
	})
}

// AddWithWeight 添加带权重的节点，权重值范围在 [1-100] 中
func (k *Ketama) AddWithWeight(node string, weight int) {
	replicas := k.replicas * weight / MaxWeight
	k.AddWithReplicas(node, replicas)
}

// Add 添加节点
func (k *Ketama) Add(node string) {
	k.AddWithReplicas(node, k.replicas)
}

// Get 获取节点
func (k *Ketama) Get(key string) (string, bool) {
	k.RLock()
	defer k.RUnlock()

	if len(k.ring) == 0 {
		return "", false
	}

	hash := k.hashFunc([]byte(key))
	index := sort.Search(len(k.ring), func(i int) bool {
		return k.ring[i] >= hash
	})

	node, ok := k.nodeMap[k.ring[index%len(k.ring)]]
	return node, ok
}

// Remove 移除节点
func (k *Ketama) Remove(nodes ...string) {
	k.Lock()
	defer k.Unlock()

	deletedHashes := make([]uint64, 0)
	for _, node := range nodes {
		for i := 0; i < k.replicas; i++ {
			hash := k.hashFunc([]byte(node + strconv.Itoa(i)))

			if _, ok := k.nodeMap[hash]; ok {
				deletedHashes = append(deletedHashes, hash)
				delete(k.nodeMap, hash)
			}
		}
	}

	if len(deletedHashes) > 0 {
		k.deleteHashes(deletedHashes)
	}
}

// deleteHashes 删除 hash 环中待删除的 hash
func (k *Ketama) deleteHashes(deletedHashes []uint64) {
	for _, hash := range deletedHashes {
		index := sort.Search(len(k.ring), func(i int) bool {
			return k.ring[i] >= hash
		})

		if index < len(k.ring) && k.ring[index] == hash {
			k.ring = append(k.ring[:index], k.ring[index+1:]...)
		}
	}
}
