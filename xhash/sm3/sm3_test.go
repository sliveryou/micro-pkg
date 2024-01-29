package sm3

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSM3_Sum(t *testing.T) {
	cases := []struct {
		data   string
		expect string
	}{
		{data: "test sm3", expect: "ea500bd4356613de09c04dbc2320566612e66d49cd7d609210074e892d07f0fa"},
		{data: "aaaaa", expect: "136ce3c86e4ed909b76082055a61586af20b4dab674732ebd4b599eef080c9be"},
		{data: "", expect: "1ab21d8355cfa17f8e61194831e81a8f22bec8c728fefb747ed035eb5082aa2b"},
	}

	for _, c := range cases {
		s := New()
		s.Write([]byte(c.data))
		got := hex.EncodeToString(s.Sum(nil))
		assert.Equal(t, c.expect, got)
	}
}

func BenchmarkSM3_Sum(b *testing.B) {
	b.ReportAllocs()
	msg := []byte("test")
	hw := New()

	for i := 0; i < b.N; i++ {
		hw.Sum(nil)
		Sum(msg)
	}
}
