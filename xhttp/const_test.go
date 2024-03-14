package xhttp

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectConstants(t *testing.T) {
	// ast 不好解析 iota 等动态常量值，此方法用于收集常量：
	// 打印输出的字符串可以用来更新下方的 headers

	thisFileName := "const_test.go"
	targetFileNames := []string{"const.go"}

	out, err := collectConstants(targetFileNames...)
	require.NoError(t, err)
	fmt.Println(out)

	in, err := os.ReadFile(thisFileName)
	require.NoError(t, err)

	content := string(in)
	reg := regexp.MustCompile(`(?s)(headers := \[]string\{.*?})`)
	gs := reg.FindStringSubmatch(content)
	if len(gs) == 2 {
		// 写入重新收集的 headers
		newContent := strings.ReplaceAll(content, gs[1], out)
		err = os.WriteFile(thisFileName, []byte(newContent), 0o666)
		require.NoError(t, err)
	}
}

func TestHeaders(t *testing.T) {
	headers := []string{
		HeaderAccept,
		HeaderAcceptLanguage,
		HeaderContentType,
		HeaderContentDisposition,
		HeaderDate,
		HeaderHost,
		HeaderLocation,
		HeaderUserAgent,
		HeaderAuthorization,
		HeaderCaErrorCode,
		HeaderCaErrorMessage,
	}

	for _, header := range headers {
		chk := http.CanonicalHeaderKey(header)
		assert.Equal(t, chk, header)
	}
}

func collectConstants(fileNames ...string) (string, error) {
	b := &strings.Builder{}
	b.WriteString("headers := []string{\n")

	for _, fileName := range fileNames {
		fSet := token.NewFileSet()
		f, err := parser.ParseFile(fSet, fileName, nil, parser.AllErrors)
		if err != nil {
			return "", errors.WithMessage(err, "parse file err")
		}

		// 解析 ast
		for _, decl := range f.Decls {
			if d, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range d.Specs {
					if s, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range s.Names {
							if name.Obj.Kind == ast.Con && strings.HasPrefix(name.Name, "Header") {
								b.WriteString(fmt.Sprintf("\t\t%s,\n", name))
							}
						}
					}
				}
			}
		}
	}
	b.WriteString("\t}")

	return b.String(), nil
}
