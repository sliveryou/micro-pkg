package bankcard4c

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	sign "github.com/sliveryou/aliyun-api-gateway-sign"
	"github.com/sliveryou/go-tool/v2/sliceg"
	"github.com/sliveryou/go-tool/v2/validator"

	"github.com/sliveryou/micro-pkg/errcode"
	"github.com/sliveryou/micro-pkg/xhttp"
)

// 阿里云银行卡四要素认证 API：https://market.aliyun.com/products/57000002/cmapi033467.html

const (
	// URL 接口地址
	URL = "https://bankcard4c.shumaidata.com/bankcard4c"

	// MsgSuccess 认证成功消息
	MsgSuccess = "认证信息匹配"
	// MsgFailure 认证失败消息
	MsgFailure = "认证失败，请稍后再试"
)

// Config 银行卡四要素认证相关配置
type Config struct {
	IsMock       bool   `json:",optional"` // 是否模拟通过
	AppKey       string `json:",optional"` // 应用Key
	AppKeySecret string `json:",optional"` // 应用密钥
}

// BankCard4C 银行卡四要素认证器结构详情
type BankCard4C struct {
	c      Config
	client *xhttp.Client
}

// NewBankCard4C 新建银行卡四要素认证器
func NewBankCard4C(c Config) (*BankCard4C, error) {
	if !c.IsMock {
		if c.AppKey == "" || c.AppKeySecret == "" {
			return nil, errors.New("bankcard4c: illegal bankcard4c config")
		}
	}

	cc := xhttp.GetDefaultConfig()
	cc.TLSHandshakeTimeout = 15 * time.Second

	return &BankCard4C{c: c, client: xhttp.NewClient(cc)}, nil
}

// MustNewBankCard4C 新建银行卡四要素认证器
func MustNewBankCard4C(c Config) *BankCard4C {
	b, err := NewBankCard4C(c)
	if err != nil {
		panic(err)
	}

	return b
}

// AuthenticateReq 银行卡四要素认证请求
type AuthenticateReq struct {
	Name     string `validate:"required" label:"姓名"`                 // 姓名
	IDCard   string `validate:"required,idcard" label:"身份证号"`        // 身份证号
	BankCard string `validate:"required,bankcard" label:"银行卡卡号"`     // 银行卡卡号
	Mobile   string `validate:"required,len=11,number" label:"电话号码"` // 电话号码
}

// AuthenticateResp 银行卡四要素认证响应
type AuthenticateResp struct {
	OrderNo string // 订单号
}

// Authenticate 银行卡四要素认证
func (b *BankCard4C) Authenticate(ctx context.Context, req *AuthenticateReq) (*AuthenticateResp, error) {
	if b.c.IsMock {
		return &AuthenticateResp{}, nil
	}

	// 校验请求参数
	if err := validator.Verify(req); err != nil {
		return nil, errcode.New(errcode.CodeInvalidParams, err.Error())
	}

	rawURL := URL
	values := make(url.Values)
	values.Set("name", req.Name)
	values.Set("idcard", req.IDCard)
	values.Set("bankcard", req.BankCard)
	values.Set("mobile", req.Mobile)
	rawURL += "?" + values.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, errors.WithMessage(err, "new http request err")
	}

	// 签名
	err = sign.Sign(request, b.c.AppKey, b.c.AppKeySecret)
	if err != nil {
		return nil, errors.WithMessage(err, "sign request err")
	}

	var resp apiResp
	response, err := b.client.CallWithRequest(request, &resp)
	if err != nil {
		return nil, errors.WithMessage(err, "client call with request err")
	}

	if resp.Code == codeSuccess && resp.Data.Result != nil {
		if *resp.Data.Result == resultConsistent {
			return &AuthenticateResp{OrderNo: resp.Data.OrderNo}, nil
		}
		return nil, errcode.NewCommon(resp.Data.Desc)
	} else if resp.Code == codeParamErr {
		return nil, errcode.NewCommon(resp.Msg)
	}

	// 获取错误消息
	messages := sliceg.Compact([]string{
		resp.Data.Desc, resp.Msg, response.Header.Get("X-Ca-Error-Message"), MsgFailure,
	})

	return nil, errors.New(messages[0])
}

const (
	// codeSuccess 接口请求成功状态码
	codeSuccess = 200
	// codeParamErr 接口参数错误状态码
	codeParamErr = 400
	// resultConsistent 认证结果：一致
	resultConsistent = 0
)

// apiResp 认证接口响应
type apiResp struct {
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Data    data   `json:"data"`
}

// data 认证接口响应数据
type data struct {
	OrderNo string `json:"order_no"`
	Result  *int   `json:"result"` // 0:一致 1:不一致 2:未认证 3:已注销
	Msg     string `json:"msg"`
	Desc    string `json:"desc"`
}
