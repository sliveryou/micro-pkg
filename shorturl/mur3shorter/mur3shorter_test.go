package mur3shorter

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShorter(t *testing.T) {
	assert.Equal(t, 61, Index)
	assert.GreaterOrEqual(t, int(math.Pow(2, Shift)), Index)
	assert.GreaterOrEqual(t, Index, int(math.Pow(2, Shift-1)))
	assert.Len(t, DefaultAlphabet, 62)

	s, err := NewShorter(6)
	require.NoError(t, err)
	assert.NotNil(t, s)

	assert.PanicsWithError(t, "murmur3: illegal shorter configure", func() {
		MustNewShorter(0)
	})
}

func TestShorter_Mapping(t *testing.T) {
	cases := []struct {
		longURL      string
		expectLength int64
	}{
		{longURL: "https://www.baidu.com", expectLength: 6},
		{longURL: "https://www.baidu.com", expectLength: 8},
		{longURL: "https://www.baidu.com", expectLength: 12},
		{longURL: "https://tieba.baidu.com", expectLength: 6},
		{longURL: "https://tieba.baidu.com", expectLength: 8},
		{longURL: "https://tieba.baidu.com", expectLength: 12},
		{longURL: "https://zhidao.baidu.com", expectLength: 6},
		{longURL: "https://zhidao.baidu.com", expectLength: 8},
		{longURL: "https://zhidao.baidu.com", expectLength: 12},
	}

	for _, c := range cases {
		s, err := NewShorter(c.expectLength)
		require.NoError(t, err)
		out := s.Mapping(c.longURL)
		assert.Len(t, out, int(c.expectLength))
		t.Log(out)
	}
}
