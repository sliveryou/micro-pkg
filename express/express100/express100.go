package express100

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/sliveryou/go-tool/v2/convert"

	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/express/types"
	"github.com/sliveryou/micro-pkg/xhash"
	"github.com/sliveryou/micro-pkg/xhttp"
)

const (
	// CloudExpress100 云服务商：快递100
	CloudExpress100 = "express100"
	// URL 接口地址
	URL = "https://poll.kuaidi100.com"

	// orderDesc 返回结果排序：降序
	orderDesc = "desc"
	// messageOK 接口请求成功消息
	messageOK = "ok"

	// comSF 快递公司编码：顺丰
	comSF = "shunfeng"
	// comSFKY 快递公司编码：顺丰快运
	comSFKY = "shunfengkuaiyun"
)

// Express100 快递100客户端结构详情
type Express100 struct {
	appID     string // 应用ID（为快递100中分配的 CustomerID）
	secretKey string // 应用密钥
	client    *xhttp.Client
}

// NewExpress100 新建快递100客户端对象
func NewExpress100(appID, secretKey string) (*Express100, error) {
	if appID == "" || secretKey == "" {
		return nil, errors.New("express100: illegal express100 config")
	}

	return &Express100{
		appID:     appID,
		secretKey: secretKey,
		client:    xhttp.NewClient(),
	}, nil
}

// Cloud 获取云服务商名称
func (e *Express100) Cloud() string {
	return CloudExpress100
}

// GetExpress 利用快递100实时快递查询接口获取快递物流信息
// https://api.kuaidi100.com/document/5f0ffb5ebc8da837cbd8aefc
func (e *Express100) GetExpress(ctx context.Context, req *types.GetExpressRequest) (*types.GetExpressResponse, error) {
	// 校验请求参数
	if req.ExpNo == "" || ((req.CoCode == comSF || req.CoCode == comSFKY) && req.TelNo == "") {
		return nil, errcode.ErrInvalidParams
	}

	// 构建请求接口地址
	rawURL := URL + "/poll/query.do"

	// 构建请求头
	header := map[string]string{
		xhttp.HeaderContentType: xhttp.ContentTypeForm,
	}

	// 构建请求参数
	param := make(map[string]string)
	param["com"] = req.CoCode  // 查询的快递公司的编码（一律用小写字母，下载编码表格：https://api.kuaidi100.com/manager/openapi/download/kdbm.do）
	param["num"] = req.ExpNo   // 查询的快递单号（单号的最小长度 6 个字符，最大长度 32 个字符）
	param["phone"] = req.TelNo // 收、寄件人的电话号码（手机和固定电话均可，只能填写一个，顺丰速运、顺丰快运必填，其他快递公司选填）
	param["order"] = orderDesc // 返回结果降序
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return nil, errors.WithMessage(err, "json marshal param err")
	}
	paramJSON := string(paramBytes)
	sign, err := xhash.HashString(md5.New(), paramJSON+e.secretKey+e.appID)
	if err != nil {
		return nil, errors.WithMessage(err, "calculate hash err")
	}

	// 构建签名参数
	values := make(url.Values)
	values.Set("customer", e.appID)
	values.Set("sign", strings.ToUpper(sign))
	values.Set("param", paramJSON)

	var resp queryResp
	err = e.client.Call(ctx, http.MethodPost, rawURL, header, strings.NewReader(values.Encode()), &resp)
	if err != nil {
		return nil, errors.WithMessage(err, "client call err")
	}

	// 判断接口请求是否成功
	if !resp.Result && resp.ReturnCode != "" && resp.Message != messageOK {
		if returnCode := convert.ToInt(resp.ReturnCode); 200 <= returnCode && returnCode <= 500 {
			return nil, errcode.NewCommon(resp.Message)
		}
		return nil, errors.New(resp.Message)
	}

	var result types.GetExpressResponse
	for _, d := range resp.Data {
		result.Traces = append(result.Traces, &types.Trace{
			Time: d.Time,
			Desc: d.Context,
		})
	}

	return &result, nil
}

// queryResp 实时快递查询响应
type queryResp struct {
	Result     bool         `json:"result"`     // 结果
	ReturnCode string       `json:"returnCode"` // 返回码
	Message    string       `json:"message"`    // 消息
	State      string       `json:"state"`      // 快递单当前状态（0-在途 1-揽收 2-疑难 3-签收 4-退签 5-派件 8-清关 14-拒签）
	Com        string       `json:"com"`        // 快递公司编码
	Nu         string       `json:"nu"`         // 单号
	Data       []*queryData `json:"data"`       // 最新查询结果
}

// queryData 实时快递查询数据
type queryData struct {
	Context string `json:"context"` // 内容
	Time    string `json:"time"`    // 时间，原始格式
	FTime   string `json:"ftime"`   // 格式化后时间
}
