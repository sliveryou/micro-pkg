package appsign

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCtxWithAppKey(t *testing.T) {
	ctx := CtxWithAppKey(context.Background(), "appkey")
	val := ctx.Value(ContextAppKey)
	assert.NotNil(t, val)
	appkey, ok := val.(string)
	assert.True(t, ok)
	assert.Equal(t, "appkey", appkey)
}

func TestAppKeyFromCtx(t *testing.T) {
	ctx := CtxWithAppKey(context.Background(), "appkey")
	appkey := AppKeyFromCtx(ctx)
	assert.Equal(t, "appkey", appkey)
}
