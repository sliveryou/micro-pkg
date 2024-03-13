package expressbird

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/express/types"
)

var (
	appID       = "appID"
	secretKey   = "secretKey"
	requestType = "1002"
	req         = &types.GetExpressRequest{
		ExpNo:  "YT12345678910",
		CoCode: "YTO",
		TelNo:  "",
	}
)

func TestNewExpressBird(t *testing.T) {
	e, err := NewExpressBird("", "", "")
	require.EqualError(t, err, "expressbird: illegal expressbird config")
	assert.Nil(t, e)

	e, err = NewExpressBird(appID, secretKey, requestType)
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, CloudExpressBird, e.Cloud())
}

func TestExpressBird_GetExpress(t *testing.T) {
	e, err := NewExpressBird(appID, secretKey, requestType)
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
