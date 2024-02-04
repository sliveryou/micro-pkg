package health

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealth(t *testing.T) {
	h := NewHealth()
	assert.True(t, h.IsUnknown())
}

func TestHealth_Up(t *testing.T) {
	h := NewHealth()
	h.Up()
	assert.True(t, h.IsUp())
}

func TestHealth_Down(t *testing.T) {
	h := NewHealth()
	h.Up()
	h.Down()
	assert.True(t, h.IsDown())
}

func TestHealth_OutOfService(t *testing.T) {
	h := NewHealth()
	h.OutOfService()
	assert.True(t, h.IsOutOfService())
}

func TestHealth_Unknown(t *testing.T) {
	h := NewHealth()
	h.Unknown()
	assert.True(t, h.IsUnknown())
}

func TestHealth_IsUp(t *testing.T) {
	h := NewHealth()
	h.Up()
	assert.Equal(t, Up, h.status)
}

func TestHealth_IsDown(t *testing.T) {
	h := NewHealth()
	h.Down()
	assert.Equal(t, Down, h.status)
}

func TestHealth_IsOutOfService(t *testing.T) {
	h := NewHealth()
	h.OutOfService()
	assert.Equal(t, OutOfService, h.status)
}

func TestHealth_IsUnknown(t *testing.T) {
	h := NewHealth()
	h.Unknown()
	assert.Equal(t, Unknown, h.status)
}

func TestHealth_MarshalJSON(t *testing.T) {
	h := NewHealth()
	h.Up()
	h.AddInfo("status", "should not render")

	j, err := h.MarshalJSON()
	require.NoError(t, err)
	expected := `{"status":"UP"}`
	assert.Equal(t, expected, string(j))

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, expected, string(b))
}

func TestHealth_UnmarshalJSON(t *testing.T) {
	h := NewHealth()
	data := `{"status":"UP","node":"node1","version":"v1.0.0"}`
	err := json.Unmarshal([]byte(data), &h)
	require.NoError(t, err)
	assert.True(t, h.IsUp())
	t.Log(h)
}

func TestHealth_AddInfo(t *testing.T) {
	h := NewHealth()
	h.AddInfo("key", "value")
	_, ok := h.info["key"]
	assert.True(t, ok)
}

func TestHealth_GetInfo(t *testing.T) {
	h := NewHealth()
	h.AddInfo("key", "value")
	value := h.GetInfo("key")
	assert.Equal(t, "value", value)
}
