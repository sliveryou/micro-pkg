package jwt

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestWithCtx(t *testing.T) {
	token := getToken()
	ctx := WithCtx(context.Background(), token)
	tok := ctx.Value(TokenKey)
	assert.NotNil(t, tok)

	tokenString, ok := tok.(string)
	assert.True(t, ok)

	newToken := &_UserInfo{}
	err := json.Unmarshal([]byte(tokenString), newToken)
	require.NoError(t, err)
	assert.Equal(t, *token, *newToken)
}

func TestReadCtx(t *testing.T) {
	err := ReadCtx(context.Background(), &struct{}{})
	require.EqualError(t, err, "no token present in context")

	token := getToken()
	ctx := WithCtx(context.Background(), token)
	newToken := &_UserInfo{}
	err = ReadCtx(ctx, newToken)
	require.NoError(t, err)
	assert.Equal(t, *token, *newToken)
}

func TestFromMD(t *testing.T) {
	token := getToken()
	tokenBytes, err := json.Marshal(token)
	require.NoError(t, err)

	md := metadata.Pairs(string(TokenKey), string(tokenBytes))
	t.Log(md)

	tokenString, ok := FromMD(md)
	assert.True(t, ok)
	assert.Equal(t, string(tokenBytes), tokenString)
	t.Log(tokenString)
}

func TestWrapAndUnwrapContext(t *testing.T) {
	token := getToken()
	tokenBytes, err := json.Marshal(token)
	require.NoError(t, err)

	pairs := []string{string(TokenKey), string(tokenBytes)}
	ctx := metadata.AppendToOutgoingContext(context.Background(), pairs...)
	md, ok := metadata.FromOutgoingContext(ctx)
	assert.True(t, ok)
	t.Log(md)

	mdToken, ok := FromMD(md)
	assert.True(t, ok)
	t.Log(mdToken)

	ctx = context.WithValue(ctx, TokenKey, mdToken)
	newToken := &_UserInfo{}
	err = ReadCtx(ctx, newToken)
	require.NoError(t, err)
	assert.Equal(t, *token, *newToken)
}
