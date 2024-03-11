package expressbird

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/sliveryou/go-tool/v2/sliceg"

	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/express/types"
	"github.com/sliveryou/micro-pkg/xhash"
	"github.com/sliveryou/micro-pkg/xhttp"
)

const (
	// CloudExpressBird 云服务商：快递鸟
	CloudExpressBird = "expressBird"
	// URL 接口地址
	URL = "https://api.kdniao.com"

	// dataTypeJSON 请求、返回数据类型均为 JSON 格式
	dataTypeJSON = "2"
	// sortDesc 轨迹排序：降序
	sortDesc = "1"
	// shipperCodeSF 快递公司编码：顺丰
	shipperCodeSF = "SF"

	requestType1002 = "1002"
	requestType8001 = "8001"
	requestType8002 = "8002"
)

// requestTypes 请求指令类型列表
var requestTypes = []string{
	"", requestType1002, requestType8001, requestType8002,
}

// ExpressBird 快递鸟客户端结构详情
type ExpressBird struct {
	appID       string // 应用ID（为快递鸟中分配的 EBusinessID）
	secretKey   string // 应用密钥
	requestType string // 请求指令类型（枚举 1002、8001 和 8002）
	client      *xhttp.Client
}

// NewExpressBird 新建快递鸟客户端对象
func NewExpressBird(appID, secretKey, requestType string) (*ExpressBird, error) {
	if appID == "" || secretKey == "" || !sliceg.Contain(requestTypes, requestType) {
		return nil, errors.New("expressbird: illegal expressbird config")
	}
	if requestType == "" {
		requestType = requestType1002
	}

	return &ExpressBird{
		appID:       appID,
		secretKey:   secretKey,
		requestType: requestType,
		client:      xhttp.NewClient(),
	}, nil
}

// Cloud 获取云服务商名称
func (e *ExpressBird) Cloud() string {
	return CloudExpressBird
}

// GetExpress 利用快递鸟即时查询接口获取快递物流信息
// 即时查询接口：https://www.kdniao.com/api-track requestType=1002，支持三家：申通、圆通和百世
// 即时查询（增值版）接口：https://www.kdniao.com/api-monitor requestType=8001
// 快递查询接口：https://www.kdniao.com/api-trackexpress requestType=8002
func (e *ExpressBird) GetExpress(ctx context.Context, req *types.GetExpressReq) (*types.GetExpressResp, error) {
	// 校验请求参数
	if req.ExpNo == "" || (req.CoCode == shipperCodeSF && req.TelNo == "") {
		return nil, errcode.ErrInvalidParams
	}

	// 构建请求接口地址
	rawURL := URL + "/Ebusiness/EbusinessOrderHandle.aspx"

	// 构建请求头
	header := map[string]string{
		xhttp.HeaderContentType: xhttp.ContentTypeForm,
	}

	// 构建请求参数
	var customerName string
	if phoneLength := len(req.TelNo); req.CoCode == shipperCodeSF && phoneLength >= 4 {
		// 处理顺丰快递查询需要的手机号信息
		customerName = req.TelNo[phoneLength-4:]
	}
	param := make(map[string]string)
	param["ShipperCode"] = req.CoCode    // 快递公司编码
	param["LogisticCode"] = req.ExpNo    // 物流单号
	param["CustomerName"] = customerName // ShipperCode 为 JD，必填，对应京东的青龙配送编码，也叫商家编码；ShipperCode 为 SF，且快递单号非快递鸟渠道返回时，必填，对应收件人/寄件人手机号后四位
	param["Sort"] = sortDesc             // 轨迹排序（0-升序 1-降序）
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return nil, errors.WithMessage(err, "json marshal param err")
	}
	paramJSON := string(paramBytes)
	sign, err := xhash.HashString(md5.New(), paramJSON+e.secretKey)
	if err != nil {
		return nil, errors.WithMessage(err, "calculate hash err")
	}

	// 构建签名参数
	values := make(url.Values)
	values.Set("RequestData", paramJSON)
	values.Set("EBusinessID", e.appID)
	values.Set("RequestType", e.requestType)
	values.Set("DataSign", base64.StdEncoding.EncodeToString([]byte(sign)))
	values.Set("DataType", dataTypeJSON)

	var resp queryResp
	err = e.client.Call(ctx, http.MethodPost, rawURL, header, strings.NewReader(values.Encode()), &resp)
	if err != nil {
		return nil, errors.WithMessage(err, "client call err")
	}

	// 判断接口请求是否成功
	if !resp.Success {
		return nil, errors.New(resp.Reason)
	}

	var result types.GetExpressResp
	for _, t := range resp.Traces {
		result.Traces = append(result.Traces, &types.Trace{
			Time: t.AcceptTime,
			Desc: t.AcceptStation,
		})
	}

	return &result, nil
}

// queryResp 快递查询结果响应
type queryResp struct {
	EBusinessID  string        `json:"EBusinessID"`  // 用户ID
	OrderCode    string        `json:"OrderCode"`    // 订单编号
	ShipperCode  string        `json:"ShipperCode"`  // 快递公司编码
	LogisticCode string        `json:"LogisticCode"` // 物流运单号
	Success      bool          `json:"Success"`      // 成功与否
	Reason       string        `json:"Reason"`       // 失败原因
	State        string        `json:"State"`        // 物流状态（0-暂无轨迹信息 1-已揽收 2-在途中 3-签收 4-问题件）
	Location     string        `json:"Location"`     // 所在城市
	Traces       []*queryTrace `json:"Traces"`       // 物流轨迹
}

// queryTrace 快递查询轨迹信息
type queryTrace struct {
	AcceptTime    string `json:"AcceptTime"`    // 时间
	AcceptStation string `json:"AcceptStation"` // 描述
	Remark        string `json:"Remark"`        // 备注
}
