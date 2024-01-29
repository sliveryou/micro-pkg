package ketama

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sliveryou/go-tool/v2/randx"
)

func TestDefaultHash(t *testing.T) {
	var hashes []uint64

	for i := 0; i < 200; i++ {
		hashes = append(hashes, DefaultHash([]byte("localhost:"+strconv.Itoa(i))))
	}

	t.Log(hashes)
	assert.Len(t, hashes, 200)
}

func TestNew(t *testing.T) {
	assert.NotNil(t, New())
	assert.NotNil(t, NewCustom(200, nil))
}

func TestKetama_Get(t *testing.T) {
	k := NewCustom(200, nil)
	k.Add("localhost:46790")
	k.AddWithWeight("localhost:46791", 100)
	k.AddWithReplicas("localhost:46792", 200)

	var l0, l1, l2 int
	for i := 0; i < 1000000; i++ {
		node, ok := k.Get(randx.NewString(64))
		if assert.True(t, ok) {
			switch node {
			case "localhost:46790":
				l0++
			case "localhost:46791":
				l1++
			case "localhost:46792":
				l2++
			}
		}
	}

	t.Log(l0, l1, l2)
	assert.Equal(t, 1000000, l0+l1+l2)
}

func TestKetama_Remove(t *testing.T) {
	k := New()
	k.Add("first")
	k.Add("second")
	k.Remove("first")

	for i := 0; i < 100; i++ {
		val, ok := k.Get(strconv.Itoa(i))
		assert.True(t, ok)
		assert.Equal(t, "second", val)
	}
}

func BenchmarkDefaultHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultHash([]byte("localhost:" + strconv.Itoa(i)))
	}
}

func BenchmarkKetama_Get(b *testing.B) {
	k := New()

	for i := 0; i < 200; i++ {
		k.Add("localhost:" + strconv.Itoa(i))
	}

	for i := 0; i < b.N; i++ {
		k.Get(strconv.Itoa(i))
	}
}

func BenchmarkKetama_Remove(b *testing.B) {
	k := New()

	for i := 0; i < 200; i++ {
		k.Add("localhost:" + strconv.Itoa(i))
	}

	for i := 0; i < b.N; i++ {
		k.Remove("localhost:" + strconv.Itoa(i))
	}
}
