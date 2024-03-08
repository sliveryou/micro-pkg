package types

// GetExpressReq 查询快递请求
type GetExpressReq struct {
	ExpNo  string // 快递单号
	CoCode string // 公司编号
	TelNo  string // 电话号码
}

// GetExpressResp 查询快递响应
type GetExpressResp struct {
	Traces []*Trace // 物流轨迹
}

// Trace 物流轨迹
type Trace struct {
	Time string // 时间
	Desc string // 描述
}
