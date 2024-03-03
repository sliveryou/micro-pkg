package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParams(t *testing.T) {
	ps := []Param{
		&CodeParam{Key: "code", Value: "123456", Length: 6, Expiration: 5 * time.Minute},
		&CommonParam{Key: "time", Value: "5"},
	}

	params := Params(ps)
	assert.Equal(t, []string{"code", "time"}, params.Keys())
	assert.Equal(t, []string{"123456", "5"}, params.Values())
	assert.Equal(t, map[string]string{"code": "123456", "time": "5"}, params.ToMap())
}
