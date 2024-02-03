package jwt

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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

func TestTokenInterceptor(t *testing.T) {
	token := getToken()
	tokenBytes, err := json.Marshal(token)
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		string(TokenKey): string(tokenBytes),
	}))

	interceptor := TokenInterceptor
	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/"}, func(ctx context.Context, req any) (any, error) {
		tok := ctx.Value(TokenKey)
		assert.NotNil(t, tok)

		tokenString, ok := tok.(string)
		assert.True(t, ok)

		newToken := &_UserInfo{}
		err := json.Unmarshal([]byte(tokenString), newToken)
		require.NoError(t, err)
		assert.Equal(t, *token, *newToken)

		return nil, err
	})
	require.NoError(t, err)
}

func TestTokenStreamInterceptor(t *testing.T) {
	token := getToken()
	tokenBytes, err := json.Marshal(token)
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		string(TokenKey): string(tokenBytes),
	}))

	interceptor := TokenStreamInterceptor
	stream := mockedStream{ctx: ctx}
	err = interceptor(nil, stream, nil, func(_ any, ss grpc.ServerStream) error {
		tok := ss.Context().Value(TokenKey)
		assert.NotNil(t, tok)

		tokenString, ok := tok.(string)
		assert.True(t, ok)

		newToken := &_UserInfo{}
		err := json.Unmarshal([]byte(tokenString), newToken)
		require.NoError(t, err)
		assert.Equal(t, *token, *newToken)

		return nil
	})
	require.NoError(t, err)
}

func TestTokenClientInterceptor(t *testing.T) {
	token := getToken()
	ctx := WithCtx(context.Background(), token)

	var run int32
	var wg sync.WaitGroup
	wg.Add(1)
	cc := new(grpc.ClientConn)
	err := TokenClientInterceptor(ctx, "/foo", nil, nil, cc,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			defer wg.Done()
			atomic.AddInt32(&run, 1)

			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			mdToken, ok := FromMD(md)
			assert.True(t, ok)

			newToken := &_UserInfo{}
			err := json.Unmarshal([]byte(mdToken), newToken)
			require.NoError(t, err)
			assert.Equal(t, *token, *newToken)

			return err
		})
	wg.Wait()
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&run))
}

func TestTokenStreamClientInterceptor(t *testing.T) {
	token := getToken()
	ctx := WithCtx(context.Background(), token)

	var run int32
	var wg sync.WaitGroup
	wg.Add(1)
	cc := new(grpc.ClientConn)
	_, err := TokenStreamClientInterceptor(ctx, nil, cc, "/foo",
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			defer wg.Done()
			atomic.AddInt32(&run, 1)

			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			mdToken, ok := FromMD(md)
			assert.True(t, ok)

			newToken := &_UserInfo{}
			err := json.Unmarshal([]byte(mdToken), newToken)
			require.NoError(t, err)
			assert.Equal(t, *token, *newToken)

			return nil, err
		})
	wg.Wait()
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&run))
}

type mockedStream struct {
	ctx context.Context
}

func (m mockedStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m mockedStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m mockedStream) SetTrailer(md metadata.MD) {
}

func (m mockedStream) Context() context.Context {
	return m.ctx
}

func (m mockedStream) SendMsg(v any) error {
	return nil
}

func (m mockedStream) RecvMsg(v any) error {
	return nil
}
