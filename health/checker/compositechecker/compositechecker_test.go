package compositechecker

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/health"
)

type upTestChecker struct{}

func (c *upTestChecker) Check(ctx context.Context) health.Health {
	h := health.NewHealth()
	h.Up()

	return h
}

type downTestChecker struct{}

func (c *downTestChecker) Check(ctx context.Context) health.Health {
	h := health.NewHealth()
	h.Down()

	return h
}

type outOfServiceTestChecker struct{}

func (c *outOfServiceTestChecker) Check(ctx context.Context) health.Health {
	h := health.NewHealth()
	h.OutOfService()

	return h
}

func TestCompositeChecker_AddChecker(t *testing.T) {
	c := NewChecker()
	assert.Empty(t, c.checkers)

	checker1 := &upTestChecker{}
	c.AddChecker("testChecker", checker1)
	require.Len(t, c.checkers, 1)
	assert.Equal(t, checker1, c.checkers[0].checker)

	checker2 := &downTestChecker{}
	c.AddChecker("testChecker", checker2)
	require.Len(t, c.checkers, 1)
	assert.NotEqual(t, checker2, c.checkers[0].checker)
}

func TestCompositeChecker_Check_Up(t *testing.T) {
	c := NewChecker()
	c.AddChecker("upTestChecker", &upTestChecker{})

	h := c.Check(context.Background())
	assert.True(t, h.IsUp())
}

func TestCompositeChecker_Check_Down(t *testing.T) {
	c := NewChecker()
	c.AddChecker("downTestChecker", &downTestChecker{})

	h := c.Check(context.Background())
	assert.True(t, h.IsDown())
}

func TestCompositeChecker_Check_OutOfService(t *testing.T) {
	c := NewChecker()
	c.AddChecker("outOfServiceTestChecker", &outOfServiceTestChecker{})

	h := c.Check(context.Background())
	assert.True(t, h.IsDown())
}

func TestCompositeChecker_Check_Down_Combined(t *testing.T) {
	c := NewChecker()
	c.AddChecker("downTestChecker", &downTestChecker{})
	c.AddChecker("upTestChecker", &upTestChecker{})

	h := c.Check(context.Background())
	assert.True(t, h.IsDown())

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"downTestChecker":{"status":"DOWN"},"status":"DOWN","upTestChecker":{"status":"UP"}}`, string(b))
}

func TestCompositeChecker_Check_Up_Combined(t *testing.T) {
	c := NewChecker()
	c.AddChecker("upTestChecker1", &upTestChecker{})
	c.AddChecker("upTestChecker2", &upTestChecker{})

	h := c.Check(context.Background())
	assert.True(t, h.IsUp())

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"UP","upTestChecker1":{"status":"UP"},"upTestChecker2":{"status":"UP"}}`, string(b))
}

func TestCompositeChecker_CheckByName(t *testing.T) {
	c := NewChecker()
	c.AddChecker("upTestChecker1", &upTestChecker{})
	c.AddChecker("upTestChecker2", &upTestChecker{})

	h := c.CheckByName(context.Background(), "unknownTestChecker")
	assert.True(t, h.IsUnknown())
	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"error":"unknown service name","status":"UNKNOWN"}`, string(b))

	h = c.CheckByName(context.Background(), "upTestChecker1")
	assert.True(t, h.IsUp())
	b, err = json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"UP"}`, string(b))

	h = c.CheckByName(context.Background(), "upTestChecker0")
	assert.True(t, h.IsUnknown())
	b, err = json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"error":"unknown service name","status":"UNKNOWN"}`, string(b))
}

func Test_CompositeChecker_AddInfo(t *testing.T) {
	c := NewChecker()

	c.AddInfo("key", "value")
	v, ok := c.info["key"]
	assert.True(t, ok)
	assert.Equal(t, "value", v)

	c.AddInfo("key", "new_value")
	v, ok = c.info["key"]
	assert.True(t, ok)
	assert.Equal(t, "new_value", v)
}
