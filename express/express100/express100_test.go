package express100

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/express/types"
)

var (
	appID     = "appID"
	secretKey = "secretKey"
	req       = &types.GetExpressReq{
		ExpNo:  "YT12345678910",
		CoCode: "yuantong",
		TelNo:  "",
	}
)

func TestNewExpress100(t *testing.T) {
	e, err := NewExpress100("", "")
	require.EqualError(t, err, "express100: illegal express100 config")
	assert.Nil(t, e)

	e, err = NewExpress100(appID, secretKey)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, CloudExpress100, e.Cloud())
}

func TestExpress100_GetExpress(t *testing.T) {
	e, err := NewExpress100(appID, secretKey)
	require.NoError(t, err)
	assert.NotNil(t, e)

	ctx := context.Background()
	resp, err := e.GetExpress(ctx, req)
	t.Log(resp, err)
	if resp != nil {
		for _, trace := range resp.Traces {
			fmt.Println(trace.Time, trace.Desc)
		}
	}
}
