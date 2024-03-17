package xinterceptor

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

	"github.com/sliveryou/micro-pkg/jwt"
)

func TestJWTInterceptor(t *testing.T) {
	token := getToken()
	tokenBytes, err := json.Marshal(token)
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		string(jwt.TokenKey): string(tokenBytes),
	}))

	interceptor := JWTInterceptor
	_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/"}, func(ctx context.Context, req any) (any, error) {
		tok := ctx.Value(jwt.TokenKey)
		assert.NotNil(t, tok)

		tokenString, ok := tok.(string)
		assert.True(t, ok)

		newToken := &userInfo{}
		err := json.Unmarshal([]byte(tokenString), newToken)
		require.NoError(t, err)
		assert.Equal(t, *token, *newToken)

		return nil, err
	})
	require.NoError(t, err)
}

func TestJWTStreamInterceptor(t *testing.T) {
	token := getToken()
	tokenBytes, err := json.Marshal(token)
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		string(jwt.TokenKey): string(tokenBytes),
	}))

	interceptor := JWTStreamInterceptor
	stream := mockedStream{ctx: ctx}
	err = interceptor(nil, stream, nil, func(_ any, ss grpc.ServerStream) error {
		tok := ss.Context().Value(jwt.TokenKey)
		assert.NotNil(t, tok)

		tokenString, ok := tok.(string)
		assert.True(t, ok)

		newToken := &userInfo{}
		err := json.Unmarshal([]byte(tokenString), newToken)
		require.NoError(t, err)
		assert.Equal(t, *token, *newToken)

		return nil
	})
	require.NoError(t, err)
}

func TestJWTClientInterceptor(t *testing.T) {
	token := getToken()
	ctx := jwt.WithCtx(context.Background(), token)

	var run int32
	var wg sync.WaitGroup
	wg.Add(1)
	cc := new(grpc.ClientConn)
	err := JWTClientInterceptor(ctx, "/foo", nil, nil, cc,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			defer wg.Done()
			atomic.AddInt32(&run, 1)

			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			mdToken, ok := jwt.FromMD(md)
			assert.True(t, ok)

			newToken := &userInfo{}
			err := json.Unmarshal([]byte(mdToken), newToken)
			require.NoError(t, err)
			assert.Equal(t, *token, *newToken)

			return err
		})
	wg.Wait()
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&run))
}

func TestJWTStreamClientInterceptor(t *testing.T) {
	token := getToken()
	ctx := jwt.WithCtx(context.Background(), token)

	var run int32
	var wg sync.WaitGroup
	wg.Add(1)
	cc := new(grpc.ClientConn)
	_, err := JWTStreamClientInterceptor(ctx, nil, cc, "/foo",
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			defer wg.Done()
			atomic.AddInt32(&run, 1)

			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			mdToken, ok := jwt.FromMD(md)
			assert.True(t, ok)

			newToken := &userInfo{}
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

func getToken() *userInfo {
	return &userInfo{
		UserID:   100000,
		UserName: "test_user",
		RoleIDs:  []int64{100000, 100001, 100002},
		Group:    "ADMIN",
		IsAdmin:  true,
		Score:    123.123,
	}
}

type userInfo struct {
	UserID   int64   `json:"user_id"`
	UserName string  `json:"user_name"`
	RoleIDs  []int64 `json:"role_ids"`
	Group    string  `json:"group"`
	IsAdmin  bool    `json:"is_admin"`
	Score    float64 `json:"score"`
}
