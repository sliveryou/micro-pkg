package face

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/collection"
)

func getFace() (*Face, error) {
	return NewFace(Config{
		IsMock:    false,
		APIKey:    "apiKey",
		SecretKey: "secretKey",
	})
}

func TestSplitN(t *testing.T) {
	cases := []struct {
		data string
		want string
	}{
		{data: "data:image/png;base64,BASE64", want: "BASE64"},
		{data: "data:video/mp4;base64,BASE64", want: "BASE64"},
		{data: "", want: ""},
		{data: " ", want: " "},
		{data: "  ", want: "  "},
		{data: ",", want: ""},
		{data: ", ", want: " "},
		{data: ",BASE64", want: "BASE64"},
		{data: "BASE64", want: "BASE64"},
	}

	for _, c := range cases {
		var s string
		got := strings.SplitN(c.data, ",", 2)
		if len(got) <= 1 {
			s = got[0]
		} else {
			s = got[1]
		}
		assert.Equal(t, c.want, s)
	}
}

func TestCache(t *testing.T) {
	cache, err := collection.NewCache(1*time.Second, collection.WithName("any"))
	require.NoError(t, err)

	cache.Set("first", "first element")
	cache.SetWithExpire("second", "second element", 1*time.Second)

	first, ok := cache.Get("first")
	assert.True(t, ok)
	assert.Equal(t, "first element", first)

	time.Sleep(2 * time.Second)
	second, ok := cache.Get("second")
	assert.False(t, ok)
	assert.Empty(t, second)

	item, err := cache.Take("third", func() (any, error) {
		return "third element", nil
	})
	require.NoError(t, err)
	third, ok := item.(string)
	assert.True(t, ok)
	assert.Equal(t, "third element", third)
}

func Test_getAccessToken(t *testing.T) {
	f, err := getFace()
	require.NoError(t, err)
	ctx := context.Background()

	token, err := f.getAccessToken(ctx)
	t.Log(token, err)
}

func TestFace_Authenticate(t *testing.T) {
	f, err := getFace()
	require.NoError(t, err)
	ctx := context.Background()

	resp, err := f.Authenticate(ctx, &AuthenticateRequest{
		Name:        "测试",
		IDCard:      "330333199001053317",
		VideoBase64: "data:video/mp4;base64,VmlkZW9CYXNlNjQ=",
	})
	t.Log(resp, err)
}
