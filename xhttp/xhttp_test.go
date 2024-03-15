package xhttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueries(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "http://test.com/api/test?id=1&hash=a&hash=b", nil)
	require.NoError(t, err)

	assert.Equal(t, "1", Query(r, "id"))
	assert.Equal(t, []string{"1"}, QueryArray(r, "id"))
	assert.Equal(t, "a", Query(r, "hash"))
	assert.Equal(t, []string{"a", "b"}, QueryArray(r, "hash"))
}

func TestParse(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://test.com/api/chunk/upload", strings.NewReader(`
    { 
		"current_data": "abcdefgh",
    	"current_seq": 1,
    	"current_size": 8,
    	"file_name": "test.txt",
    	"file_hash": "ec3f5c9819f41ec8965587553fbe9935ec26ec440c5adc94ff6c10efadeba80f",
    	"total_seq": 1,
    	"total_size": 8 
	}`))
	require.NoError(t, err)
	r.Header.Set(HeaderContentType, MIMEApplicationJSON)

	req := &_UploadFileChunkReq{}
	err = Parse(r, req)
	require.NoError(t, err)
	assert.Equal(t, "abcdefgh", req.CurrentData)
	assert.Equal(t, int32(1), req.CurrentSeq)
	assert.Equal(t, int32(8), req.CurrentSize)
	assert.Equal(t, "test.txt", req.FileName)
	assert.Equal(t, "ec3f5c9819f41ec8965587553fbe9935ec26ec440c5adc94ff6c10efadeba80f", req.FileHash)
	assert.Equal(t, int64(1), req.TotalSeq)
	assert.Equal(t, int64(8), req.TotalSize)
}

func TestParseForm(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet,
		"http://test.com/api/test?file_type=txt&file_hashes=a&file_hashes=b&file_seqs=1&file_seqs=2", nil)
	require.NoError(t, err)

	req1 := _GetFilesReq{}
	err = ParseForm(r, &req1)
	require.NoError(t, err)
	assert.Equal(t, "txt", req1.FileType)
	assert.Equal(t, []string{"a", "b"}, req1.FileHashes)
	assert.Equal(t, []int64{1, 2}, req1.FileSeqs)

	r, err = http.NewRequest(http.MethodPost, "http://test.com/api/test",
		strings.NewReader("file_type=txt&file_hashes=a&file_hashes=b&file_seqs=1&file_seqs=2"))
	require.NoError(t, err)
	r.Header.Set(HeaderContentType, MIMEForm)

	req2 := _GetFilesReq{}
	err = ParseForm(r, &req2)
	require.NoError(t, err)
	assert.Equal(t, "txt", req2.FileType)
	assert.Equal(t, []string{"a", "b"}, req2.FileHashes)
	assert.Equal(t, []int64{1, 2}, req2.FileSeqs)
}

func TestParseJsonBody(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://test.com/api/chunk/upload", strings.NewReader(`
    { 
		"current_data": "abcdefgh",
    	"current_seq": 1,
    	"current_size": 8,
    	"file_name": "test.txt",
    	"file_hash": "ec3f5c9819f41ec8965587553fbe9935ec26ec440c5adc94ff6c10efadeba80f",
    	"total_seq": 1,
    	"total_size": 8 
	}`))
	require.NoError(t, err)
	r.Header.Set(HeaderContentType, MIMEApplicationJSON)

	req := &_UploadFileChunkReq{}
	err = ParseJsonBody(r, req)
	require.NoError(t, err)
	assert.Equal(t, "abcdefgh", req.CurrentData)
	assert.Equal(t, int32(1), req.CurrentSeq)
	assert.Equal(t, int32(8), req.CurrentSize)
	assert.Equal(t, "test.txt", req.FileName)
	assert.Equal(t, "ec3f5c9819f41ec8965587553fbe9935ec26ec440c5adc94ff6c10efadeba80f", req.FileHash)
	assert.Equal(t, int64(1), req.TotalSeq)
	assert.Equal(t, int64(8), req.TotalSize)
}

func TestFromFile(t *testing.T) {
	fileAbsName := "../testdata/test.txt"
	fileReader, err := os.Open(fileAbsName)
	require.NoError(t, err)
	fileName := path.Base(fileAbsName)

	body := &bytes.Buffer{}
	writer := NewMultipartWriter(body)

	err = writer.WriteField("file_type", "txt")
	require.NoError(t, err)

	ct, err := writer.WriteFile("file_data", fileName, fileReader)
	require.NoError(t, err)
	require.Equal(t, "text/plain", ct)

	contentType := writer.FormDataContentType()
	err = writer.Close()
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodPost, "http://test.com/api/file/upload", body)
	r.Header.Set(HeaderContentType, contentType)
	w := httptest.NewRecorder()
	testFileHandler(w, r)

	result := w.Result()
	defer result.Body.Close()

	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	data := string(d)
	assert.Equal(t, "{\"code\":0,\"msg\":\"ok\",\"data\":{\"file_data\":\"A test text.\\nA test text.\\nA test text.\\n\",\"file_name\":\"test.txt\",\"file_size\":39,\"file_type\":\"txt\"}}", data)
}

func testFileHandler(w http.ResponseWriter, r *http.Request) {
	type _UploadFileReq struct {
		FileType string `form:"file_type" validate:"required" label:"文件类型"` // 文件类型
	}

	ctx := r.Context()
	var req _UploadFileReq
	if err := Parse(r, &req); err != nil {
		ErrorCtx(ctx, w, err)
		return
	}

	fh, err := FromFile(r, "file_data")
	if err != nil {
		ErrorCtx(ctx, w, err)
		return
	}

	f, err := fh.Open()
	if err != nil {
		ErrorCtx(ctx, w, err)
		return
	}
	defer f.Close()

	d, err := io.ReadAll(f)
	if err != nil {
		ErrorCtx(ctx, w, err)
		return
	}

	OkJsonCtx(ctx, w, map[string]any{
		"file_type": req.FileType,
		"file_data": string(d),
		"file_name": fh.Filename,
		"file_size": fh.Size,
	})
}

func TestGetClientIP(t *testing.T) {
	cases := []struct {
		k      string
		v      string
		expect string
	}{
		{k: "X-Forwarded-For", v: "127.0.0.1", expect: "127.0.0.1"},
		{k: "X-Real-Ip", v: "127.0.0.1", expect: "127.0.0.1"},
		{k: "X-Appengine-Remote-Addr", v: "127.0.0.1", expect: "127.0.0.1"},
		{k: "Test", v: "127.0.0.1", expect: ""},
		{k: "", v: "", expect: ""},
	}

	for _, c := range cases {
		r, err := http.NewRequest(http.MethodGet, "http://test.com/api/test?id=1&hash=a&hash=b", nil)
		require.NoError(t, err)
		r.Header.Set(c.k, c.v)
		assert.Equal(t, c.expect, GetClientIP(r))
	}
}

func TestGetInternalIP(t *testing.T) {
	iip := GetInternalIP()
	assert.NotEmpty(t, iip)
	t.Log(iip)
}

func TestCopyRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "http://test.com/api/chunk/upload", strings.NewReader(`
    { 
		"current_data": "abcdefgh",
    	"current_seq": 1,
    	"current_size": 8,
    	"file_name": "test.txt",
    	"file_hash": "ec3f5c9819f41ec8965587553fbe9935ec26ec440c5adc94ff6c10efadeba80f",
    	"total_seq": 1,
    	"total_size": 8
	}`))
	r.Header.Set(HeaderContentType, MIMEApplicationJSON)
	w := httptest.NewRecorder()
	testHandler(w, r)

	result := w.Result()
	defer result.Body.Close()

	d, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	data := string(d)
	assert.Equal(t, "{\"code\":0,\"msg\":\"ok\",\"data\":{\"is_equal\":true,\"req1\":{\"current_data\":\"abcdefgh\",\"current_seq\":1,\"current_size\":8,\"file_name\":\"test.txt\",\"file_hash\":\"ec3f5c9819f41ec8965587553fbe9935ec26ec440c5adc94ff6c10efadeba80f\",\"total_seq\":1,\"total_size\":8},\"req2\":{\"current_data\":\"abcdefgh\",\"current_seq\":1,\"current_size\":8,\"file_name\":\"test.txt\",\"file_hash\":\"ec3f5c9819f41ec8965587553fbe9935ec26ec440c5adc94ff6c10efadeba80f\",\"total_seq\":1,\"total_size\":8}}}", data)
}

func TestErrorCtx1(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "http://test.com/api/chunk/upload", strings.NewReader(`
    { 
		"current_data": 1234,
    	"current_seq": 1,
    	"current_size": 8,
    	"file_name": "test.txt",
    	"file_hash": "ec3f5c9819f41ec8965587553fbe9935ec26ec440c5adc94ff6c10efadeba80f",
    	"total_seq": 1,
    	"total_size": 8
	}`))
	r.Header.Set(HeaderContentType, MIMEApplicationJSON)
	w := httptest.NewRecorder()
	testHandler(w, r)

	result := w.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, "{\"code\":100,\"msg\":\"请求参数错误\"}", string(data))
}

func TestErrorCtx2(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "http://test.com/api/chunk/upload", strings.NewReader(`
    { 
		"current_data": "",
    	"current_seq": 1,
    	"current_size": 8,
    	"file_name": "test.txt",
    	"file_hash": "ec3f5c9819f41ec8965587553fbe9935ec26ec440c5adc94ff6c10efadeba80f",
    	"total_seq": 1,
    	"total_size": 8
	}`))
	r.Header.Set(HeaderContentType, MIMEApplicationJSON)
	w := httptest.NewRecorder()
	testHandler(w, r)

	result := w.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	assert.Equal(t, "{\"code\":100,\"msg\":\"当前块数据为必填字段\"}", string(data))
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rc, err := CopyRequest(r)
	if err != nil {
		ErrorCtx(ctx, w, err)
		return
	}

	var req1 _UploadFileChunkReq
	if err := Parse(r, &req1); err != nil {
		ErrorCtx(ctx, w, err)
		return
	}

	var req2 _UploadFileChunkReq
	if err := Parse(rc, &req2); err != nil {
		ErrorCtx(ctx, w, err)
		return
	}

	OkJsonCtx(ctx, w, map[string]any{
		"req1":     req1,
		"req2":     req2,
		"is_equal": req1 == req2,
	})
}

type _GetFilesReq struct {
	FileType   string   `form:"file_type" validate:"required" label:"文件类型"`                       // 文件类型
	FileHashes []string `form:"file_hashes" validate:"required,dive,required" label:"文件 hash 列表"` // 文件 hash 列表
	FileSeqs   []int64  `form:"file_seqs" validate:"required,dive,required" label:"文件序号列表"`       // 文件序号列表
}

type _UploadFileChunkReq struct {
	CurrentData string `json:"current_data" validate:"required" label:"当前块数据"`                  // 当前块数据（须 base64 编码）
	CurrentSeq  int32  `json:"current_seq" validate:"required" label:"当前块序号"`                   // 当前块序号（从 1 开始）
	CurrentSize int32  `json:"current_size" validate:"required" label:"当前块大小"`                  // 当前块大小
	FileName    string `json:"file_name" validate:"required" label:"文件名"`                       // 文件名
	FileHash    string `json:"file_hash" validate:"required,len=64,hexadecimal" label:"文件hash"` // 文件hash（sha256(文件内容+文件名称)）
	TotalSeq    int64  `json:"total_seq" validate:"required" label:"总序号"`                       // 总序号
	TotalSize   int64  `json:"total_size" validate:"required" label:"总大小"`                      // 总大小
}
