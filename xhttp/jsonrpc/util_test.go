package jsonrpc

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParams(t *testing.T) {
	ps := Params(person{})
	assert.Equal(t, reflect.Struct, reflect.TypeOf(ps).Kind())

	ps = Params(map[string]any{"a": "b"})
	assert.Equal(t, reflect.Map, reflect.TypeOf(ps).Kind())

	ps = Params(person{}, &person{})
	assert.Equal(t, reflect.Slice, reflect.TypeOf(ps).Kind())

	ps = Params(1)
	assert.Equal(t, reflect.Slice, reflect.TypeOf(ps).Kind())

	ps = Params("abc")
	assert.Equal(t, reflect.Slice, reflect.TypeOf(ps).Kind())
}

func Test_unmarshal(t *testing.T) {
	body := []byte(`{"name":"sliveryou","age":18,"country":"China"}`)
	m1 := make(map[string]any)
	err := unmarshal(body, &m1)
	require.NoError(t, err)

	name, ok := m1["name"].(string)
	assert.True(t, ok)
	assert.Equal(t, "sliveryou", name)

	ageNumber, ok := m1["age"].(json.Number)
	assert.True(t, ok)
	age, err := ageNumber.Int64()
	require.NoError(t, err)
	assert.Equal(t, int64(18), age)

	country, ok := m1["country"].(string)
	assert.True(t, ok)
	assert.Equal(t, "China", country)

	m2 := make(map[string]any)
	err = json.Unmarshal(body, &m2)
	require.NoError(t, err)

	ageFloat, ok := m2["age"].(float64)
	assert.True(t, ok)
	assert.InDelta(t, float64(18), ageFloat, 0.001)
}
