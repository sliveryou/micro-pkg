package shorturl

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/go-tool/v2/randx"

	"github.com/sliveryou/micro-pkg/shorturl/mur3shorter"
)

func TestNewShorter(t *testing.T) {
	s, err := NewShorter(Config{
		Length: 6,
	})
	require.NoError(t, err)
	assert.NotNil(t, s)

	assert.PanicsWithError(t, "murmur3: illegal shorter configure", func() {
		MustNewShorter(Config{
			Length: 0,
		})
	})
}

func TestShorter_Mapping(t *testing.T) {
	cases := []struct {
		longURL      string
		expectLength int64
	}{
		{longURL: "https://www.baidu.com", expectLength: 6},
		{longURL: "https://www.baidu.com", expectLength: 8},
		{longURL: "https://www.baidu.com", expectLength: 10},
		{longURL: "https://tieba.baidu.com", expectLength: 6},
		{longURL: "https://tieba.baidu.com", expectLength: 8},
		{longURL: "https://tieba.baidu.com", expectLength: 10},
		{longURL: "https://zhidao.baidu.com", expectLength: 6},
		{longURL: "https://zhidao.baidu.com", expectLength: 8},
		{longURL: "https://zhidao.baidu.com", expectLength: 10},
	}

	for _, c := range cases {
		s, err := NewShorter(Config{
			Length: c.expectLength,
		})
		require.NoError(t, err)
		out := s.Mapping(c.longURL)
		assert.Len(t, out, int(c.expectLength))
		t.Log(out)
	}
}

func TestShorter_Mapping_InCurrency(t *testing.T) {
	s := MustNewShorter(Config{
		Length:   6,
		Alphabet: mur3shorter.DefaultAlphabet,
	})

	c := make(chan string)
	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			c <- s.Mapping(randx.NewString(100))
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	r := make([]string, 0, 100)
	for s := range c {
		r = append(r, s)
	}

	uniqueMap := make(map[string]struct{}, len(r))
	for _, s := range r {
		uniqueMap[s] = struct{}{}
	}
	isUnique := len(uniqueMap) == len(r)
	assert.True(t, isUnique)
}
