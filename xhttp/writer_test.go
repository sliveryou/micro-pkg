package xhttp

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMultipartWriter(t *testing.T) {
	fileAbsName := "test/test.txt"
	fileReader, err := os.Open(fileAbsName)
	require.NoError(t, err)
	fileName := filepath.Base(fileAbsName)

	fields := url.Values{}
	fields.Add("ext", path.Ext(fileName))
	fields.Add("name", fileName)
	fields.Add("key", "key")
	fields.Add("OSSAccessKeyId", "OSSAccessKeyId")
	fields.Add("policy", "policy")
	fields.Add("callback", "callback")
	fields.Add("signature", "signature")
	fields.Add("success_action_status", "200")

	body := &bytes.Buffer{}
	writer := NewMultipartWriter(body)

	for fieldName := range fields {
		err := writer.WriteField(fieldName, fields.Get(fieldName))
		require.NoError(t, err)
	}
	ct, err := writer.WriteFile("file", fileName, fileReader)
	require.NoError(t, err)
	require.Equal(t, "text/plain; charset=utf-8", ct)

	contentType := writer.FormDataContentType()
	err = writer.Close()
	require.NoError(t, err)
	require.Contains(t, contentType, "multipart/form-data; boundary=")

	fmt.Println(contentType)
	fmt.Println(ct)
	fmt.Println(body.String())
}

func TestMultipartWriter_WriteFile(t *testing.T) {
	fileAbsName := "test/test.pdf"
	fileReader, err := os.Open(fileAbsName)
	require.NoError(t, err)
	fileName := filepath.Base(fileAbsName)

	body := &bytes.Buffer{}
	writer := NewMultipartWriter(body)

	ct, err := writer.WriteFile("file", fileName, fileReader)
	require.NoError(t, err)
	require.Equal(t, "application/pdf", ct)

	contentType := writer.FormDataContentType()
	err = writer.Close()
	require.NoError(t, err)
	require.Contains(t, contentType, "multipart/form-data; boundary=")

	fmt.Println(contentType)
	fmt.Println(ct)
}
