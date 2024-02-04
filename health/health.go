package health

import (
	"context"
	"encoding/json"
)

//go:generate enumer -type Status -json -linecomment -output health_string.go

// Status 健康状态
type Status int32

const (
	// Up 健康状态：UP
	Up Status = 0 // UP
	// Down 健康状态：DOWN
	Down Status = 1 // DOWN
	// OutOfService 健康状态：OUT OF SERVICE
	OutOfService Status = 2 // OUT OF SERVICE
	// Unknown 健康状态：UNKNOWN
	Unknown Status = 3 // UNKNOWN
)

// Health 健康状态结构详情
type Health struct {
	status Status
	info   map[string]any
}

// Checker 应用检查器接口
type Checker interface {
	// Check 检查应用健康情况
	Check(ctx context.Context) Health
}

// NewHealth 新建健康状态，默认状态为 Unknown
func NewHealth() Health {
	return Health{
		status: Unknown,
		info:   make(map[string]any),
	}
}

// MarshalJSON 实现 json.Marshaler 接口的 MarshalJSON 方法
func (h Health) MarshalJSON() ([]byte, error) {
	data := make(map[string]any)

	for k, v := range h.info {
		data[k] = v
	}
	data["status"] = h.status

	return json.Marshal(data)
}

// GetInfo 获取健康键值信息
func (h Health) GetInfo(key string) any {
	return h.info[key]
}

// GetStatus 获取健康状态信息
func (h Health) GetStatus() Status {
	return h.status
}

// IsUp 判断是否为 Up 状态
func (h Health) IsUp() bool {
	return h.status == Up
}

// IsDown 判断是否为 Down 状态
func (h Health) IsDown() bool {
	return h.status == Down
}

// IsOutOfService 判断是否为 IsOutOfService 状态
func (h Health) IsOutOfService() bool {
	return h.status == OutOfService
}

// IsUnknown 判断是否为 Unknown 状态
func (h Health) IsUnknown() bool {
	return h.status == Unknown
}

// UnmarshalJSON 实现 json.Unmarshaler 接口的 UnmarshalJSON 方法
func (h *Health) UnmarshalJSON(data []byte) error {
	v := make(map[string]any)
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	h.status = Unknown
	if s, ok := v["status"]; ok {
		if ss, ok := s.(string); ok {
			if status, err := StatusString(ss); err == nil {
				h.status = status
			}
		}
	}
	delete(v, "status")
	h.info = v

	return nil
}

// AddInfo 添加健康键值信息
func (h *Health) AddInfo(key string, value any) *Health {
	h.info[key] = value
	return h
}

// Up 设置健康状态为 Up
func (h *Health) Up() *Health {
	h.status = Up
	return h
}

// Down 设置健康状态为 Down
func (h *Health) Down() *Health {
	h.status = Down
	return h
}

// OutOfService 设置健康状态为 OutOfService
func (h *Health) OutOfService() *Health {
	h.status = OutOfService
	return h
}

// Unknown 设置健康状态为 Unknown
func (h *Health) Unknown() *Health {
	h.status = Unknown
	return h
}
