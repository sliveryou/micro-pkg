package corpaccount

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	sign "github.com/sliveryou/aliyun-api-gateway-sign"
	"github.com/sliveryou/go-tool/v2/convert"
	"github.com/sliveryou/go-tool/v2/sliceg"
	"github.com/sliveryou/go-tool/v2/validator"

	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/xhttp"
)

// 阿里云企业银行卡账户认证 API：https://market.aliyun.com/products/57000002/cmapi027344.html

const (
	// URL 接口地址
	URL = "https://verifycorp.market.alicloudapi.com/lianzhuo/verifyCorpAccount"

	// MsgSuccess 认证成功消息
	MsgSuccess = "认证信息匹配"
	// MsgFailure 认证失败消息
	MsgFailure = "认证失败，请稍后再试"
)

// MockTransAmt 模拟打款金额，单位（分）
var MockTransAmt = 5

// Config 企业银行卡账户认证相关配置
type Config struct {
	IsMock       bool   // 是否模拟通过
	AppKey       string // 应用key
	AppKeySecret string // 应用密钥
}

// CorpAccount 企业银行卡账户认证器结构详情
type CorpAccount struct {
	c      Config
	client *xhttp.Client
}

// NewCorpAccount 新建企业银行卡账户认证器
func NewCorpAccount(c Config) (*CorpAccount, error) {
	if !c.IsMock {
		if c.AppKey == "" || c.AppKeySecret == "" {
			return nil, errors.New("corpaccount: illegal corpaccount config")
		}
	}

	cc := xhttp.GetDefaultConfig()
	cc.HTTPTimeout = 60 * time.Second
	cc.TLSHandshakeTimeout = 15 * time.Second

	return &CorpAccount{c: c, client: xhttp.NewClient(cc)}, nil
}

// MustNewCorpAccount 新建企业银行卡账户认证器
func MustNewCorpAccount(c Config) *CorpAccount {
	b, err := NewCorpAccount(c)
	if err != nil {
		panic(err)
	}

	return b
}

// AuthenticateReq 企业银行卡账户认证请求
type AuthenticateReq struct {
	CardNo   string `validate:"required,corpaccount" label:"企业账号"` // 企业账号
	AcctName string `validate:"required" label:"企业名称"`             // 企业名称
	BankName string `validate:"required" label:"开户行名称"`            // 开户行名称（要和银行列表里名称完全一致：http://lundroid.com/basedata/3.xlsx?spm=5176.product-detail.detail.7.5d11386fyiwyaW&file=3.xlsx）
}

// AuthenticateResp 企业银行卡账户认证响应
type AuthenticateResp struct {
	RequestNo int64  // 请求编号
	TransAmt  int    // 打款随机金额，单位（分）
	Abstract  string // 打款摘要
}

// Authenticate 企业银行卡账户认证
func (c *CorpAccount) Authenticate(ctx context.Context, req *AuthenticateReq) (*AuthenticateResp, error) {
	if c.c.IsMock {
		return &AuthenticateResp{TransAmt: MockTransAmt}, nil
	}

	// 校验请求参数
	if err := validator.Verify(req); err != nil {
		return nil, errcode.New(errcode.CodeInvalidParams, err.Error())
	}

	rawURL := URL
	values := make(url.Values)
	values.Set("cardno", req.CardNo)
	values.Set("acctName", req.AcctName)
	values.Set("bankName", req.BankName)
	rawURL += "?" + values.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, errors.WithMessage(err, "new http request err")
	}

	// 签名
	err = sign.Sign(request, c.c.AppKey, c.c.AppKeySecret)
	if err != nil {
		return nil, errors.WithMessage(err, "sign request err")
	}

	var resp apiResp
	response, err := c.client.CallWithRequest(request, &resp)
	if err != nil {
		return nil, errors.WithMessage(err, "client call with request err")
	}

	if resp.Code != nil {
		if code := *resp.Code; code == codeSuccess {
			return &AuthenticateResp{
				RequestNo: resp.RequestNo,
				TransAmt:  convert.ToInt(resp.TransAmt),
				Abstract:  resp.Abstract,
			}, nil
		} else if errMsg, ok := errMap[code]; ok {
			return nil, errcode.NewCommon(errMsg)
		}
	}

	// 获取错误消息
	messages := sliceg.Compact([]string{
		resp.Desc, response.Header.Get("X-Ca-Error-Message"), MsgFailure,
	})

	return nil, errors.New(messages[0])
}

const (
	// codeSuccess 接口请求成功状态码
	codeSuccess = 0
)

// apiResp 认证接口响应
type apiResp struct {
	Code      *int   `json:"code"`      // 错误码
	Desc      string `json:"desc"`      // 错误信息
	RequestNo int64  `json:"requestNo"` // 请求编号
	TransAmt  string `json:"transamt"`  // 打款随机金额，单位（分）
	Abstract  string `json:"abstract"`  // 打款摘要
}

var errMap = map[int]string{
	2: "账号与开户名不符",
	3: "开户行名称错误",
	4: "仅支持对公账户验证",
}
