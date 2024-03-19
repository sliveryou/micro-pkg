package xreq

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	stdurl "net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions_errors(t *testing.T) {
	t.Run("Custom Option error", func(t *testing.T) {
		_, err := NewPost("https://www.test.com/api", customOption{})
		require.EqualError(t, err, "option collection apply customOption err: custom option apply err")
	})

	t.Run("URL error", func(t *testing.T) {
		_, err := NewPost("https://www.test.com/api", URL("%"))
		require.EqualError(t, err, `option collection apply URL err: parse url: % err: parse "%": invalid URL escape "%"`)
	})

	t.Run("BodyJSON error", func(t *testing.T) {
		_, err := NewPost("https://www.test.com/api", BodyJSON(make(chan struct{})))
		require.ErrorContains(t, err, "option collection apply BodyJSON err: json marshal obj")
		require.ErrorContains(t, err, "json: unsupported type: chan struct {}")
	})

	t.Run("BodyXML error", func(t *testing.T) {
		_, err := NewPost("https://www.test.com/api", BodyXML(make(chan struct{})))
		require.ErrorContains(t, err, "option collection apply BodyXML err: xml marshal obj")
		require.ErrorContains(t, err, "xml: unsupported type: chan struct {}")
	})

	t.Run("Dump read error", func(t *testing.T) {
		_, err := NewGet("https://www.test.com/api",
			BodyReader(errReadWriter{}),
			Dump(os.Stdout),
		)
		require.EqualError(t, err, "option collection apply Dump err: dump request err: read err")
	})

	t.Run("Dump write error", func(t *testing.T) {
		_, err := NewGet("https://www.test.com/api",
			Dump(errReadWriter{}),
		)
		require.EqualError(t, err, "option collection apply Dump err: write dump err: write err")
	})
}

func TestDump(t *testing.T) {
	var buffer bytes.Buffer
	_, err := New(http.MethodPost, "http://www.test.com/api",
		Scheme("https"),
		Host("my.test.com"),
		AddPath("/students"),
		BearerAuth("abcdefgh"),
		BodyJSON(map[string]string{"name": "SliverYou", "language": "go"}),
		Dump(&buffer),
	)
	require.NoError(t, err)

	expected := "POST /api/students HTTP/1.1\r\n" +
		"Host: my.test.com\r\n" +
		"Authorization: Bearer abcdefgh\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		"{\"language\":\"go\",\"name\":\"SliverYou\"}"

	assert.Equal(t, expected, buffer.String())
}

func TestParseURL(t *testing.T) {
	url := "https://www.test.com/api?a=1&b=2&c=3&c=4#ok"
	u, err := stdurl.Parse(url)
	require.NoError(t, err)
	fmt.Println(u)

	assert.Equal(t, "https", u.Scheme)
	assert.Equal(t, "www.test.com", u.Host)
	assert.Equal(t, "/api", u.Path)
	assert.Empty(t, u.RawPath)
	assert.Equal(t, stdurl.Values{"a": {"1"}, "b": {"2"}, "c": {"3", "4"}}, u.Query())
	assert.Equal(t, "a=1&b=2&c=3&c=4", u.RawQuery)
	assert.Equal(t, "ok", u.Fragment)
	assert.Empty(t, u.RawFragment)
}

func Test_nopCloser_Size(t *testing.T) {
	cases := []struct {
		reader     io.Reader
		expectSize int64
	}{
		{reader: bytes.NewBufferString("test reader"), expectSize: 11},
		{reader: bytes.NewReader([]byte("test reader")), expectSize: 11},
		{reader: strings.NewReader("test reader"), expectSize: 11},
		{reader: io.LimitReader(strings.NewReader("test reader"), 20), expectSize: 20},
		{reader: io.NewSectionReader(strings.NewReader("test reader"), 0, 20), expectSize: 20},
		{reader: nil, expectSize: 0},
	}

	for _, c := range cases {
		got := rc(c.reader).Size()
		require.Equal(t, c.expectSize, got)
	}
}

type customOption struct{}

func (customOption) Apply(request *http.Request) (*http.Request, error) {
	return nil, errors.New("custom option apply err")
}

type errReadWriter struct{}

func (errReadWriter) Read([]byte) (int, error) {
	return 0, errors.New("read err")
}

func (errReadWriter) Write([]byte) (int, error) {
	return 0, errors.New("write err")
}
