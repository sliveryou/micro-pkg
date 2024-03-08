package xhash

import (
	"hash"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashReader(t *testing.T) {
	cases := []struct {
		h          hash.Hash
		r          io.Reader
		expectHash string
	}{
		{h: New("md5"), r: strings.NewReader("test md5"), expectHash: "0e4e3b2681e8931c067a23c583c878d5"},
		{h: New("sm3"), r: strings.NewReader("test sm3"), expectHash: "ea500bd4356613de09c04dbc2320566612e66d49cd7d609210074e892d07f0fa"},
		{h: New("sha1"), r: strings.NewReader("test sha1"), expectHash: "b99c071333d4dbca0d9298e5c8d7480f176cafdc"},
		{h: New("sha256"), r: strings.NewReader("test sha256"), expectHash: "c71d137da140c5afefd7db8e7a255df45c2ac46064e934416dc04020a91f3fd2"},
		{h: New("sha512"), r: strings.NewReader("test sha512"), expectHash: "247bd141d3b8a886ebe1f60fa62d9ff60ffdc33efc43e8d76d24af4b02324ebec2b3fc55c5d0da7f9bd5ef8536b3d2056ec6ff5d452d7e4bc0b77cdb66fafc85"},
	}

	for _, c := range cases {
		s, err := HashReader(c.h, c.r)
		require.NoError(t, err)
		assert.Equal(t, c.expectHash, s)
	}
}

func TestBase64HashReader(t *testing.T) {
	cases := []struct {
		h          hash.Hash
		r          io.Reader
		expectHash string
	}{
		{h: New("md5"), r: strings.NewReader("test md5"), expectHash: "Dk47JoHokxwGeiPFg8h41Q=="},
		{h: New("sm3"), r: strings.NewReader("test sm3"), expectHash: "6lAL1DVmE94JwE28IyBWZhLmbUnNfWCSEAdOiS0H8Po="},
		{h: New("sha1"), r: strings.NewReader("test sha1"), expectHash: "uZwHEzPU28oNkpjlyNdIDxdsr9w="},
		{h: New("sha256"), r: strings.NewReader("test sha256"), expectHash: "xx0TfaFAxa/v19uOeiVd9FwqxGBk6TRBbcBAIKkfP9I="},
		{h: New("sha512"), r: strings.NewReader("test sha512"), expectHash: "JHvRQdO4qIbr4fYPpi2f9g/9wz78Q+jXbSSvSwIyTr7Cs/xVxdDaf5vV74U2s9IFbsb/XUUtfkvAt3zbZvr8hQ=="},
	}

	for _, c := range cases {
		s, err := Base64HashReader(c.h, c.r)
		require.NoError(t, err)
		assert.Equal(t, c.expectHash, s)
	}
}

func TestHashString(t *testing.T) {
	cases := []struct {
		h          hash.Hash
		s          string
		expectHash string
	}{
		{h: New("md5"), s: "test md5", expectHash: "0e4e3b2681e8931c067a23c583c878d5"},
		{h: New("sm3"), s: "test sm3", expectHash: "ea500bd4356613de09c04dbc2320566612e66d49cd7d609210074e892d07f0fa"},
		{h: New("sha1"), s: "test sha1", expectHash: "b99c071333d4dbca0d9298e5c8d7480f176cafdc"},
		{h: New("sha256"), s: "test sha256", expectHash: "c71d137da140c5afefd7db8e7a255df45c2ac46064e934416dc04020a91f3fd2"},
		{h: New("sha512"), s: "test sha512", expectHash: "247bd141d3b8a886ebe1f60fa62d9ff60ffdc33efc43e8d76d24af4b02324ebec2b3fc55c5d0da7f9bd5ef8536b3d2056ec6ff5d452d7e4bc0b77cdb66fafc85"},
	}

	for _, c := range cases {
		s, err := HashString(c.h, c.s)
		require.NoError(t, err)
		assert.Equal(t, c.expectHash, s)
	}
}

func TestBase64HashString(t *testing.T) {
	cases := []struct {
		h          hash.Hash
		s          string
		expectHash string
	}{
		{h: New("md5"), s: "test md5", expectHash: "Dk47JoHokxwGeiPFg8h41Q=="},
		{h: New("sm3"), s: "test sm3", expectHash: "6lAL1DVmE94JwE28IyBWZhLmbUnNfWCSEAdOiS0H8Po="},
		{h: New("sha1"), s: "test sha1", expectHash: "uZwHEzPU28oNkpjlyNdIDxdsr9w="},
		{h: New("sha256"), s: "test sha256", expectHash: "xx0TfaFAxa/v19uOeiVd9FwqxGBk6TRBbcBAIKkfP9I="},
		{h: New("sha512"), s: "test sha512", expectHash: "JHvRQdO4qIbr4fYPpi2f9g/9wz78Q+jXbSSvSwIyTr7Cs/xVxdDaf5vV74U2s9IFbsb/XUUtfkvAt3zbZvr8hQ=="},
	}

	for _, c := range cases {
		s, err := Base64HashString(c.h, c.s)
		require.NoError(t, err)
		assert.Equal(t, c.expectHash, s)
	}
}

func TestHashFile(t *testing.T) {
	cases := []struct {
		h          hash.Hash
		filePath   string
		fileName   string
		expectHash string
	}{
		{h: New("md5"), filePath: "../testdata/test.pdf", fileName: "", expectHash: "1c8e7cef13cad4b8a633bbfd159b12de"},
		{h: New("md5"), filePath: "../testdata/test.pdf", fileName: "test.pdf", expectHash: "464a54cf4c07dcef9c3877f6b4a45232"},
		{h: New("sm3"), filePath: "../testdata/test.pdf", fileName: "", expectHash: "78138b5ca8314d1bdafcc9324d3fb525865f8acba0af7ddde2aa0ad1d205704d"},
		{h: New("sm3"), filePath: "../testdata/test.pdf", fileName: "test.pdf", expectHash: "7ac867afcdcf487006acf633382f7f7b184e169d7659c3043a3c6b02c9397cbe"},
		{h: New("sha256"), filePath: "../testdata/test.pdf", fileName: "", expectHash: "36307dc0d069ba6e2b9beb9d17b052210c837f7d35c41b1bbd9fc18dd767e786"},
		{h: New("sha256"), filePath: "../testdata/test.pdf", fileName: "test.pdf", expectHash: "0e72ca723ab43bfd90509436c01e3db184066b5cb1bf28e13fe84fb9e6e576a3"},
	}

	for _, c := range cases {
		s, err := HashFile(c.h, c.filePath, c.fileName)
		require.NoError(t, err)
		assert.Equal(t, c.expectHash, s)
	}
}

func TestBcrypt(t *testing.T) {
	cases := []struct {
		pwd string
	}{
		{pwd: "123123"},
		{pwd: "123456"},
		{pwd: "ABCDEFGH12345678"},
		{pwd: "Abc123"},
	}

	for _, c := range cases {
		hashedPwd, err := GenFromPwd(c.pwd)
		require.NoError(t, err)
		assert.True(t, CmpHashAndPwd(hashedPwd, c.pwd))
		t.Log(hashedPwd, c.pwd)
	}
}
