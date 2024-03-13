package express

import (
	"context"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/express/express100"
	"github.com/sliveryou/micro-pkg/express/expressbird"
	"github.com/sliveryou/micro-pkg/express/types"
)

// Express 快递查询客户端接口
type Express interface {
	// Cloud 获取云服务商名称
	Cloud() string
	// GetExpress 获取快递物流信息
	GetExpress(ctx context.Context, req *types.GetExpressRequest) (*types.GetExpressResponse, error)
}

// Config 快递查询客户端相关配置
type Config struct {
	Cloud       string `json:",options=[express100,expressBird]"` // 云服务商（当前支持 express100 和 expressBird）
	AppID       string // 应用ID
	SecretKey   string // 应用密钥
	RequestType string `json:",optional"` // 请求指令类型（expressBird 云服务商专用，枚举 1002、8001 和 8002）
}

// NewExpress 新建快递查询客户端对象
func NewExpress(c Config) (Express, error) {
	var client Express
	var err error

	switch c.Cloud {
	case express100.CloudExpress100:
		client, err = express100.NewExpress100(c.AppID, c.SecretKey)
	case expressbird.CloudExpressBird:
		client, err = expressbird.NewExpressBird(c.AppID, c.SecretKey, c.RequestType)
	default:
		err = errors.New("illegal express cloud")
	}
	if err != nil {
		return nil, errors.WithMessage(err, "express: new express client err")
	}

	return &defaultExpress{c: c, client: client}, nil
}

// MustNewExpress 新建快递查询客户端对象
func MustNewExpress(c Config) Express {
	e, err := NewExpress(c)
	if err != nil {
		panic(err)
	}

	return e
}

// defaultExpress 默认快递查询客户端结构详情
type defaultExpress struct {
	c      Config
	client Express
}

// Cloud 获取云服务商名称
func (e *defaultExpress) Cloud() string {
	return e.client.Cloud()
}

// GetExpress 获取快递物流信息
func (e *defaultExpress) GetExpress(ctx context.Context, req *types.GetExpressRequest) (*types.GetExpressResponse, error) {
	return e.client.GetExpress(ctx, req)
}
