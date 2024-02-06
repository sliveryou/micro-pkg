package xhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// source from:
	//   https://raw.githubusercontent.com/jshttp/mime-db/master/db.json
	mimeDBFile = "../testdata/mime-db.json"
	reg        = regexp.MustCompile(`(?s)(var extToMimeType = map\[string]string\{.*?\n})`)
)

func TestParseMineType(t *testing.T) {
	type rawMineType struct {
		Source       string   `json:"source"`
		Compressible bool     `json:"compressible"`
		Extensions   []string `json:"extensions"`
	}

	type mimeType struct {
		MimeType string `json:"mime_type"`
		Source   string `json:"source"`
		Priority int    `json:"priority"`
	}

	type extension struct {
		Extension string     `json:"extension"`
		MimeTypes []mimeType `json:"mime_types"`
	}

	m := make(map[string]rawMineType)
	mimeDB, err := os.ReadFile(mimeDBFile)
	require.NoError(t, err)
	err = json.Unmarshal(mimeDB, &m)
	require.NoError(t, err)

	mts := make([]string, 0, len(m))
	for mt := range m {
		mts = append(mts, mt)
	}
	sort.Strings(mts)

	var exts []string
	em := make(map[string]extension)
	for _, mt := range mts {
		mti := m[mt]
		for _, ext := range mti.Extensions {
			ei, ok := em[ext]
			if !ok {
				ei.Extension = "." + ext
				exts = append(exts, ext)
			}

			var priority int
			switch mti.Source {
			case "iana":
				priority = 3
			case "apache":
				priority = 2
			case "nginx":
				priority = 1
			}

			ei.MimeTypes = append(ei.MimeTypes, mimeType{MimeType: mt, Source: mti.Source, Priority: priority})
			sort.Slice(ei.MimeTypes, func(i, j int) bool {
				if ei.MimeTypes[i].Priority != ei.MimeTypes[j].Priority {
					return ei.MimeTypes[i].Priority > ei.MimeTypes[j].Priority
				}
				return len(ei.MimeTypes[i].MimeType) < len(ei.MimeTypes[j].MimeType)
			})
			em[ext] = ei
		}
	}
	sort.Strings(exts)

	var b strings.Builder
	b.WriteString("var extToMimeType = map[string]string{\n")
	for _, ext := range exts {
		ei := em[ext]
		if ei.MimeTypes[0].MimeType == ApplicationStream && len(ei.MimeTypes) > 1 {
			ei.MimeTypes[0], ei.MimeTypes[1] = ei.MimeTypes[1], ei.MimeTypes[0]
		}
		b.WriteString(fmt.Sprintf("\t%q: %q,\n", ei.Extension, ei.MimeTypes[0].MimeType))
	}
	b.WriteString("}")
	o, err := format.Source([]byte(b.String()))
	require.NoError(t, err)
	out := string(o)
	fmt.Println(out)

	in, err := os.ReadFile("util.go")
	require.NoError(t, err)

	content := string(in)
	gs := reg.FindStringSubmatch(content)
	if len(gs) == 2 {
		// 写入重新收集的 extToMimeType
		newContent := strings.ReplaceAll(content, gs[1], out)
		err = os.WriteFile("util.go", []byte(newContent), 0o666)
		require.NoError(t, err)
	}
}

func TestTypeByExtension(t *testing.T) {
	cases := []struct {
		filePath string
		expect   string
	}{
		{filePath: "test.txt", expect: "text/plain"},
		{filePath: "test.pdf", expect: "application/pdf"},
		{filePath: "test.jpg", expect: "image/jpeg"},
		{filePath: "test", expect: "application/octet-stream"},
		{filePath: "/root/dir/test.txt", expect: "text/plain"},
		{filePath: "test.txt", expect: "text/plain"},
		{filePath: "root/dir/test.txt", expect: "text/plain"},
		{filePath: "root\\dir\\test.txt", expect: "text/plain"},
		{filePath: "D:\\work\\dir\\test.txt", expect: "text/plain"},
	}

	for _, c := range cases {
		got := TypeByExtension(c.filePath)
		assert.Contains(t, got, c.expect)
	}
}

func TestGetReaderLen(t *testing.T) {
	cases := []struct {
		reader    io.Reader
		expectLen int64
		expectErr bool
	}{
		{reader: bytes.NewBufferString("test reader"), expectLen: 11, expectErr: false},
		{reader: bytes.NewReader([]byte("test reader")), expectLen: 11, expectErr: false},
		{reader: strings.NewReader("test reader"), expectLen: 11, expectErr: false},
		{reader: io.LimitReader(strings.NewReader("test reader"), 20), expectLen: 20, expectErr: false},
		{reader: io.NewSectionReader(strings.NewReader("test reader"), 0, 20), expectLen: 20, expectErr: false},
		{reader: &_mockReader{}, expectLen: 0, expectErr: true},
	}

	for _, c := range cases {
		got, err := GetReaderLen(c.reader)
		if c.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, c.expectLen, got)
		}
	}
}

func TestParseEndpoint(t *testing.T) {
	cases := []struct {
		endpoint string
		isUseSSL bool
	}{
		{endpoint: "", isUseSSL: false},
		{endpoint: ":", isUseSSL: false},
		{endpoint: ":56789", isUseSSL: false},
		{endpoint: "56789", isUseSSL: false},
		{endpoint: "0.0.0.0:56789/bucket", isUseSSL: false},
		{endpoint: "localhost:56789/bucket", isUseSSL: false},
		{endpoint: "172.0.0.100:56789/bucket", isUseSSL: false},
		{endpoint: "http://0.0.0.0:56789/bucket", isUseSSL: false},
		{endpoint: "https://localhost:56789/bucket", isUseSSL: true},
		{endpoint: "https://172.0.0.100:56789/bucket", isUseSSL: true},
	}

	for _, c := range cases {
		pe, useSSL := ParseEndpoint(c.endpoint)
		assert.False(t, strings.HasPrefix(pe, "http"))
		assert.Equal(t, c.isUseSSL, useSSL)
		fmt.Println(pe)
	}
}

type _mockReader struct{}

func (m *_mockReader) Read(p []byte) (n int, err error) {
	return 0, nil
}
