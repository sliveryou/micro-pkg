package gstream

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStreamWriter(t *testing.T) {
	type testReq1 struct {
		FileIndex int64
		FileData  []byte
	}
	_, err := NewStreamWriter(nil, &testReq1{FileData: nil}, "FileData", (3<<20)+(1<<19))
	require.NoError(t, err)

	type testReq2 struct {
		fileData []byte
	}
	_, err = NewStreamWriter(nil, &testReq2{fileData: nil}, "fileData", (3<<20)+(1<<19))
	require.EqualError(t, err, "gstream: field should be exported")

	_, err = NewStreamWriter(nil, nil, "FileData", (3<<20)+(1<<19))
	require.EqualError(t, err, "gstream: target cannot be nil")

	_, err = NewStreamWriter(nil, &testReq1{FileData: nil}, "FileDataX", (3<<20)+(1<<19))
	require.EqualError(t, err, "gstream: field should be target struct field")

	_, err = NewStreamWriter(nil, &testReq1{FileData: nil}, "FileIndex", (3<<20)+(1<<19))
	require.EqualError(t, err, "gstream: field should be []byte type")
}

func TestNewStreamReader(t *testing.T) {
	type testReq1 struct {
		FileIndex int64
		FileData  []byte
	}
	_, err := NewStreamReader(nil, &testReq1{FileData: nil}, "FileData", (3<<20)+(1<<19))
	require.NoError(t, err)

	type testReq2 struct {
		fileData []byte
	}
	_, err = NewStreamReader(nil, &testReq2{fileData: nil}, "fileData", (3<<20)+(1<<19))
	require.EqualError(t, err, "gstream: field should be exported")

	_, err = NewStreamReader(nil, nil, "FileData", (3<<20)+(1<<19))
	require.EqualError(t, err, "gstream: target cannot be nil")

	_, err = NewStreamReader(nil, &testReq1{FileData: nil}, "FileDataX", (3<<20)+(1<<19))
	require.EqualError(t, err, "gstream: field should be target struct field")

	_, err = NewStreamReader(nil, &testReq1{FileData: nil}, "FileIndex", (3<<20)+(1<<19))
	require.EqualError(t, err, "gstream: field should be []byte type")
}

func TestReflect(t *testing.T) {
	type testReq struct {
		FileData []byte
	}

	v, err := checkAndGetTargetValue(&testReq{FileData: nil}, "FileData")
	require.NoError(t, err)

	vt := v.Type()
	t.Log(vt)

	target := reflect.New(vt).Elem()
	t.Log(target)

	target.FieldByName("FileData").SetBytes([]byte("test set bytes"))
	t.Log(target, target.Addr().Interface())

	testFunc := func(r any) {
		req, ok := r.(*testReq)
		if assert.True(t, ok) {
			t.Log(string(req.FileData))
		}
	}

	testFunc(target.Addr().Interface())

	v, err = checkAndGetTargetValue(testReq{FileData: nil}, "FileData")
	require.NoError(t, err)
	assert.NotNil(t, v)

	var a *testReq
	v, err = checkAndGetTargetValue(a, "FileData")
	require.EqualError(t, err, "gstream: target value is invalid")
	assert.NotNil(t, v)
}
