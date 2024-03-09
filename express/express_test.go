package express

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/express/express100"
	"github.com/sliveryou/micro-pkg/express/expressbird"
	"github.com/sliveryou/micro-pkg/express/types"
)

func TestMustNewExpress(t *testing.T) {
	assert.PanicsWithError(t, "express: new express client err: illegal express cloud", func() {
		c := Config{
			Cloud: "unknown",
		}
		MustNewExpress(c)
	})

	assert.PanicsWithError(t, "express: new express client err: expressbird: illegal expressbird config", func() {
		c := Config{
			Cloud: expressbird.CloudExpressBird,
			AppID: "appID",
		}
		MustNewExpress(c)
	})
}

func TestExpress100_GetExpress(t *testing.T) {
	var (
		c = Config{
			Cloud:     express100.CloudExpress100,
			AppID:     "appID",
			SecretKey: "secretKey",
		}
		req = &types.GetExpressReq{
			ExpNo:  "YT12345678910",
			CoCode: "yuantong",
			TelNo:  "",
		}
	)

	e, err := NewExpress(c)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, express100.CloudExpress100, e.Cloud())

	ctx := context.Background()
	resp, err := e.GetExpress(ctx, req)
	t.Log(resp, err)
	if resp != nil {
		for _, trace := range resp.Traces {
			fmt.Println(trace.Time, trace.Desc)
		}
	}
}

func TestExpressBird_GetExpress(t *testing.T) {
	var (
		c = Config{
			Cloud:       expressbird.CloudExpressBird,
			AppID:       "appID",
			SecretKey:   "secretKey",
			RequestType: "1002",
		}
		req = &types.GetExpressReq{
			ExpNo:  "YT12345678910",
			CoCode: "YTO",
			TelNo:  "",
		}
	)

	e, err := NewExpress(c)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, expressbird.CloudExpressBird, e.Cloud())

	ctx := context.Background()
	resp, err := e.GetExpress(ctx, req)
	t.Log(resp, err)
	if resp != nil {
		for _, trace := range resp.Traces {
			fmt.Println(trace.Time, trace.Desc)
		}
	}
}
