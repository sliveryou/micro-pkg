package enforcer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/threading"
)

func TestNewEnforcer(t *testing.T) {
	e, err := NewEnforcer(Config{}, &MockAdapter{}, &MockWatcher{})
	require.NoError(t, err)
	assert.Equal(t, DefaultModelText, e.c.ModelText)
	assert.Equal(t, 500*time.Millisecond, e.c.RetryDuration)
	assert.Equal(t, 5, e.c.RetryMaxTimes)
}

func TestMustNewEnforcer(t *testing.T) {
	assert.NotPanics(t, func() {
		e := MustNewEnforcer(Config{}, &MockAdapter{}, &MockWatcher{})
		assert.Equal(t, DefaultModelText, e.c.ModelText)
		assert.Equal(t, 500*time.Millisecond, e.c.RetryDuration)
		assert.Equal(t, 5, e.c.RetryMaxTimes)
	})
}

func TestEnforcer_Enforce(t *testing.T) {
	e := MustNewEnforcer(Config{}, &MockAdapter{}, &MockWatcher{})

	cases := []struct {
		sub  string
		obj  string
		act  string
		isOK bool
	}{
		{sub: "ADMIN", obj: "/api/department", act: "GET", isOK: true},
		{sub: "ADMIN", obj: "/api/job/1", act: "DELETE", isOK: true},
		{sub: "ADMIN", obj: "/api/personnel/1", act: "DELETE", isOK: true},
		{sub: "ADMIN", obj: "/api/job", act: "DELETE", isOK: false},
		{sub: "NOT_ADMIN", obj: "/api/personnel/1", act: "GET", isOK: false},
	}

	for _, c := range cases {
		got, err := e.Enforce(c.sub, c.obj, c.act)
		require.NoError(t, err)
		assert.Equal(t, c.isOK, got)
	}
}

func TestEnforcer_Update(t *testing.T) {
	e := MustNewEnforcer(Config{}, &MockAdapter{}, &MockWatcher{})

	e.Update()

	time.Sleep(2 * time.Second)
}

func TestEnforcer_Reload(t *testing.T) {
	e := MustNewEnforcer(Config{}, &MockAdapter{}, &MockWatcher{})

	threading.GoSafe(func() {
		e.Reload(500*time.Millisecond, 3)
	})
	threading.GoSafe(func() {
		e.Reload(500*time.Millisecond, 3)
	})
	threading.GoSafe(func() {
		e.Reload(500*time.Millisecond, 3)
	})

	time.Sleep(2 * time.Second)
}
