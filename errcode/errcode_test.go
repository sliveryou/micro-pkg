package errcode

import (
	stderrors "errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/go-tool/v2/convert"
	"github.com/sliveryou/go-tool/v2/sliceg"
)

func TestCollectConstants(t *testing.T) {
	// ast 不好解析 iota 等动态常量值，此方法用于收集常量：
	// 打印输出的字符串可以用来更新下方的 constantsMap

	thisFileName := "errcode_test.go"
	targetFileNames := []string{"errcode.go"}

	out, err := collectConstants(targetFileNames...)
	require.NoError(t, err)
	fmt.Println(out)

	in, err := os.ReadFile(thisFileName)
	require.NoError(t, err)

	content := string(in)
	reg := regexp.MustCompile(`(?s)(var constantsMap = map\[string]any\{.*?})`)
	gs := reg.FindStringSubmatch(content)
	if len(gs) == 2 {
		// 写入重新收集的 constantsMap
		newContent := strings.ReplaceAll(content, gs[1], out)
		err = os.WriteFile(thisFileName, []byte(newContent), 0o666)
		require.NoError(t, err)
	}
}

var constantsMap = map[string]any{
	"CodeOK":             CodeOK,
	"MsgOK":              MsgOK,
	"CodeCommon":         CodeCommon,
	"MsgCommon":          MsgCommon,
	"CodeRecordNotFound": CodeRecordNotFound,
	"MsgRecordNotFound":  MsgRecordNotFound,
	"CodeUnexpected":     CodeUnexpected,
	"MsgUnexpected":      MsgUnexpected,
	"CodeInvalidParams":  CodeInvalidParams,
	"MsgInvalidParams":   MsgInvalidParams,
	"GrpcMaxCode":        GrpcMaxCode,
}

func TestGenDoc(t *testing.T) {
	// 执行该测试用例，可以生成 errcode.md 文件
	// 执行前请确保 constantsMap 已更新至最新，否则请执行 TestCollectConstants 测试用例
	b := &strings.Builder{}
	b.WriteString("# 接口错误码\n\n" +
		"> 注意：HTTP 状态码为 `200`，错误码为 `0` 时，才代表接口请求成功\n\n" +
		"| **错误** | **错误码** | **释义** | **HTTP 状态码** |\n" +
		"|:---------|:-----------|:---------|:----------------|\n")

	fileNames := []string{"errcode.go"}
	targetFileName := "../docs/errcode.md"

	out, err := genDoc(fileNames...)
	require.NoError(t, err)
	b.WriteString(out)

	err = os.WriteFile(targetFileName, []byte(b.String()), 0o666)
	require.NoError(t, err)
}

func TestNew(t *testing.T) {
	cases := []struct {
		code           uint32
		msg            string
		httpCode       int
		expectHTTPCode int
	}{
		{code: 101, msg: "101 error", httpCode: 200, expectHTTPCode: 200},
		{code: 102, msg: "102 error", httpCode: 0, expectHTTPCode: 200},
		{code: 103, msg: "103 error", httpCode: 301, expectHTTPCode: 301},
	}

	for _, c := range cases {
		e := New(c.code, c.msg, c.httpCode)
		require.Error(t, e)

		var err *Err
		ok := errors.As(e, &err)

		assert.True(t, ok)
		assert.NotNil(t, err)
		assert.Equal(t, c.code, err.Code)
		assert.Equal(t, c.msg, err.Msg)
		assert.Equal(t, c.expectHTTPCode, err.HTTPCode)
	}
}

func TestNewCommon(t *testing.T) {
	cases := []struct {
		msg            string
		httpCode       int
		expectHTTPCode int
	}{
		{msg: "101 error", httpCode: 200, expectHTTPCode: 200},
		{msg: "102 error", httpCode: 0, expectHTTPCode: 200},
		{msg: "103 error", httpCode: 301, expectHTTPCode: 301},
	}

	for _, c := range cases {
		e := NewCommon(c.msg, c.httpCode)
		require.Error(t, e)

		var err *Err
		ok := errors.As(e, &err)

		assert.True(t, ok)
		assert.NotNil(t, err)
		assert.Equal(t, uint32(CodeCommon), err.Code)
		assert.Equal(t, c.msg, err.Msg)
		assert.Equal(t, c.expectHTTPCode, err.HTTPCode)
	}
}

func TestErr_Error(t *testing.T) {
	cases := []struct {
		code        uint32
		msg         string
		httpCode    int
		expectError string
	}{
		{code: 101, msg: "101 error", httpCode: 200, expectError: "101 error"},
		{code: 102, msg: "102 error", httpCode: 0, expectError: "102 error"},
		{code: 103, msg: "103 error", httpCode: 301, expectError: "103 error"},
	}

	for _, c := range cases {
		e := New(c.code, c.msg, c.httpCode)
		require.Error(t, e)

		var err *Err
		ok := errors.As(e, &err)

		assert.True(t, ok)
		assert.NotNil(t, err)
		assert.Equal(t, c.expectError, err.Error())
	}
}

func TestErr_String(t *testing.T) {
	cases := []struct {
		code         uint32
		msg          string
		httpCode     int
		expectString string
	}{
		{code: 101, msg: "101 error", httpCode: 200, expectString: "code: 101, msg: 101 error, http code: 200"},
		{code: 102, msg: "102 error", httpCode: 0, expectString: "code: 102, msg: 102 error, http code: 200"},
		{code: 103, msg: "103 error", httpCode: 301, expectString: "code: 103, msg: 103 error, http code: 301"},
	}

	for _, c := range cases {
		e := New(c.code, c.msg, c.httpCode)
		require.Error(t, e)

		var err *Err
		ok := errors.As(e, &err)

		assert.True(t, ok)
		assert.NotNil(t, err)
		assert.Equal(t, c.expectString, err.String())
	}
}

func TestErr_GRPCStatus(t *testing.T) {
	cases := []struct {
		code         uint32
		msg          string
		httpCode     int
		expectString string
	}{
		{code: 101, msg: "101 error", httpCode: 200},
		{code: 102, msg: "102 error", httpCode: 0},
		{code: 103, msg: "103 error", httpCode: 301},
	}

	for _, c := range cases {
		e := New(c.code, c.msg, c.httpCode)
		require.Error(t, e)

		var err *Err
		ok := errors.As(e, &err)

		assert.True(t, ok)
		assert.NotNil(t, err)

		grpcStatus := err.GRPCStatus()
		assert.NotNil(t, grpcStatus)
		assert.Equal(t, uint32(grpcStatus.Code()), err.Code)
		assert.Equal(t, grpcStatus.Message(), err.String())
		if uint32(grpcStatus.Code()) > GrpcMaxCode {
			assert.Equal(t, grpcStatus.Code().String(), fmt.Sprintf("Code(%d)", err.Code))
		}
	}
}

func TestFromMessage(t *testing.T) {
	cases := []struct {
		message   string
		expectErr error
		expectOK  bool
	}{
		{message: "code: 101, msg: 101 error, http code: 200", expectErr: New(101, "101 error"), expectOK: true},
		{message: "code: 102, msg: 102 error, http code: 200", expectErr: New(102, "102 error"), expectOK: true},
		{message: "", expectErr: ErrUnexpected, expectOK: false},
		{message: "test message", expectErr: ErrUnexpected, expectOK: false},
	}

	for _, c := range cases {
		err, ok := FromMessage(c.message)
		assert.Equal(t, c.expectOK, ok)
		assert.NotNil(t, err)
		assert.True(t, IsErr(err))
		assert.True(t, Is(err, c.expectErr))
	}
}

func TestFromError(t *testing.T) {
	cases := []struct {
		err    error
		expect bool
	}{
		{err: OK, expect: true},
		{err: ErrUnexpected, expect: true},
		{err: ErrInvalidParams, expect: true},
		{err: errors.WithMessage(ErrInvalidParams, "test wrap"), expect: true},
		{err: nil, expect: true},
		{err: New(CodeUnexpected, MsgUnexpected), expect: true},
		{err: errors.New("test"), expect: false},
		{err: stderrors.New("test"), expect: false},
		{err: New(101, "101 error"), expect: true},
	}

	for _, c := range cases {
		err, ok := FromError(c.err)
		assert.Equal(t, c.expect, ok)
		require.Error(t, err)
		assert.True(t, IsErr(err))
	}
}

func TestIs(t *testing.T) {
	cases := []struct {
		err    error
		target error
		expect bool
	}{
		{err: New(CodeOK, MsgOK), target: OK, expect: true},
		{err: New(CodeUnexpected, MsgUnexpected), target: ErrUnexpected, expect: true},
		{err: ErrInvalidParams, target: ErrUnexpected, expect: false},
		{err: New(CodeUnexpected, MsgUnexpected), target: ErrInvalidParams, expect: false},
		{err: New(CodeUnexpected, MsgUnexpected), target: New(CodeUnexpected, MsgUnexpected), expect: true},
		{err: New(CodeInvalidParams, MsgInvalidParams), target: ErrInvalidParams, expect: true},
		{err: ErrInvalidParams, target: errors.New("错误"), expect: false},
		{err: errors.New("错误"), target: ErrUnexpected, expect: false},
	}

	for _, c := range cases {
		assert.Equal(t, c.expect, Is(c.err, c.target))
	}
}

func genDoc(fileNames ...string) (string, error) {
	b := &strings.Builder{}
	rowMap := make(map[int64]docRow)

	var docRows []docRow
	for _, fileName := range fileNames {
		drs, err := genDocRows(fileName)
		if err != nil {
			return "", err
		}
		docRows = append(docRows, drs...)
	}

	if len(docRows) > 0 {
		sliceg.SortFunc(docRows, func(a, b docRow) int { return int(a.Code - b.Code) })
		for _, dr := range docRows {
			if r, ok := rowMap[dr.Code]; ok {
				return "", errors.Errorf("业务错误码冲突：%d，已存在的相同错误码的业务错误: %+v，冲突的业务错误：%+v", dr.Code, r, dr)
			}
			rowMap[dr.Code] = dr
			color := "green"
			if dr.HTTPCode != 200 {
				color = "red"
			}

			b.WriteString(fmt.Sprintf("| %s | %d | %s | <font color='%s'>%d</font> |\n",
				dr.Name, dr.Code, dr.Msg, color, dr.HTTPCode))
		}
	}

	return b.String(), nil
}

func collectConstants(fileNames ...string) (string, error) {
	b := &strings.Builder{}
	b.WriteString("var constantsMap = map[string]any{\n")

	for _, fileName := range fileNames {
		fSet := token.NewFileSet()
		f, err := parser.ParseFile(fSet, fileName, nil, parser.AllErrors)
		if err != nil {
			return "", err
		}

		// 解析 ast
		for _, decl := range f.Decls {
			if d, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range d.Specs {
					if s, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range s.Names {
							if name.Obj.Kind == ast.Con {
								b.WriteString(fmt.Sprintf("\t%q: %s,\n", name, name))
							}
						}
					}
				}
			}
		}
	}
	b.WriteString("}")

	out, err := format.Source([]byte(b.String()))
	if err != nil {
		return "", err
	}

	return string(out), nil
}

type docRow struct {
	Name     string
	Code     int64
	Msg      string
	HTTPCode int64
}

func (dc *docRow) isEmpty() bool {
	return dc.Name == "" || dc.Msg == "" || dc.HTTPCode == 0
}

func genDocRows(fileName string) ([]docRow, error) {
	fSet := token.NewFileSet()
	f, err := parser.ParseFile(fSet, fileName, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	// 解析 ast
	var docRows []docRow
	for _, decl := range f.Decls {
		if d, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range d.Specs {
				if s, ok := spec.(*ast.ValueSpec); ok {
					if len(s.Names) != len(s.Values) {
						continue
					}
					for i := 0; i < len(s.Names); i++ {
						dr := docRow{Name: s.Names[i].Name, HTTPCode: 200}
						if v, ok := s.Values[i].(*ast.CallExpr); ok && len(v.Args) >= 2 {
							// 判断是否为 New 方法
							isCodeNew := false
							if vf, ok := v.Fun.(*ast.Ident); ok && vf.Name == "New" {
								isCodeNew = true
							}
							// 判断是否为 errcode.New 方法
							if vf, ok := v.Fun.(*ast.SelectorExpr); ok && vf.Sel.Name == "New" {
								if vx, ok := vf.X.(*ast.Ident); ok && vx.Name == "errcode" {
									isCodeNew = true
								}
							}
							if !isCodeNew {
								continue
							}
							for j, arg := range v.Args {
								col := ""
								switch a := arg.(type) {
								case *ast.BasicLit:
									col = strings.Trim(a.Value, `"`)
								case *ast.Ident:
									if decl, ok := a.Obj.Decl.(*ast.ValueSpec); ok && len(decl.Values) > 0 {
										if dv, ok := decl.Values[0].(*ast.BasicLit); ok {
											col = strings.Trim(dv.Value, `"`)
										}
									}
									if col == "" {
										if con, ok := constantsMap[a.Name]; ok {
											col = convert.ToString(con)
										}
									}
								case *ast.SelectorExpr:
									col = convertStatusCode(fmt.Sprintf("%s.%s", a.X, a.Sel.Name))
								}
								if col != "" {
									if j == 0 {
										dr.Code = convert.ToInt64(col)
									} else if j == 1 {
										dr.Msg = col
									} else if j == 2 {
										dr.HTTPCode = convert.ToInt64(col)
									}
								}
							}
						}
						// else {
						// 	// 包外错误需要硬编码
						// 	var e *errcode.Err
						// 	switch dr.Name {
						// 	case "InvalidParams":
						// 		errors.As(errcode.ErrInvalidParams, &e)
						// 	case "Unexpected":
						// 		errors.As(errcode.ErrUnexpected, &e)
						// 	}
						// 	if e != nil {
						// 		dr = docRow{Name: dr.Name, Code: int64(e.Code), Msg: e.Msg, HTTPCode: int64(e.HTTPCode)}
						// 	}
						// }
						if !dr.isEmpty() {
							docRows = append(docRows, dr)
						}
					}
				}
			}
		}
	}

	return docRows, nil
}

func convertStatusCode(name string) string {
	statusCodeMap := map[string]string{
		"http.StatusContinue":                      "100",
		"http.StatusSwitchingProtocols":            "101",
		"http.StatusProcessing":                    "102",
		"http.StatusEarlyHints":                    "103",
		"http.StatusOK":                            "200",
		"http.StatusCreated":                       "201",
		"http.StatusAccepted":                      "202",
		"http.StatusNonAuthoritativeInfo":          "203",
		"http.StatusNoContent":                     "204",
		"http.StatusResetContent":                  "205",
		"http.StatusPartialContent":                "206",
		"http.StatusMultiStatus":                   "207",
		"http.StatusAlreadyReported":               "208",
		"http.StatusIMUsed":                        "226",
		"http.StatusMultipleChoices":               "300",
		"http.StatusMovedPermanently":              "301",
		"http.StatusFound":                         "302",
		"http.StatusSeeOther":                      "303",
		"http.StatusNotModified":                   "304",
		"http.StatusUseProxy":                      "305",
		"http.StatusTemporaryRedirect":             "307",
		"http.StatusPermanentRedirect":             "308",
		"http.StatusBadRequest":                    "400",
		"http.StatusUnauthorized":                  "401",
		"http.StatusPaymentRequired":               "402",
		"http.StatusForbidden":                     "403",
		"http.StatusNotFound":                      "404",
		"http.StatusMethodNotAllowed":              "405",
		"http.StatusNotAcceptable":                 "406",
		"http.StatusProxyAuthRequired":             "407",
		"http.StatusRequestTimeout":                "408",
		"http.StatusConflict":                      "409",
		"http.StatusGone":                          "410",
		"http.StatusLengthRequired":                "411",
		"http.StatusPreconditionFailed":            "412",
		"http.StatusRequestEntityTooLarge":         "413",
		"http.StatusRequestURITooLong":             "414",
		"http.StatusUnsupportedMediaType":          "415",
		"http.StatusRequestedRangeNotSatisfiable":  "416",
		"http.StatusExpectationFailed":             "417",
		"http.StatusTeapot":                        "418",
		"http.StatusMisdirectedRequest":            "421",
		"http.StatusUnprocessableEntity":           "422",
		"http.StatusLocked":                        "423",
		"http.StatusFailedDependency":              "424",
		"http.StatusTooEarly":                      "425",
		"http.StatusUpgradeRequired":               "426",
		"http.StatusPreconditionRequired":          "428",
		"http.StatusTooManyRequests":               "429",
		"http.StatusRequestHeaderFieldsTooLarge":   "431",
		"http.StatusUnavailableForLegalReasons":    "451",
		"http.StatusInternalServerError":           "500",
		"http.StatusNotImplemented":                "501",
		"http.StatusBadGateway":                    "502",
		"http.StatusServiceUnavailable":            "503",
		"http.StatusGatewayTimeout":                "504",
		"http.StatusHTTPVersionNotSupported":       "505",
		"http.StatusVariantAlsoNegotiates":         "506",
		"http.StatusInsufficientStorage":           "507",
		"http.StatusLoopDetected":                  "508",
		"http.StatusNotExtended":                   "510",
		"http.StatusNetworkAuthenticationRequired": "511",
	}

	if code, ok := statusCodeMap[name]; ok {
		return code
	}

	return "200"
}
