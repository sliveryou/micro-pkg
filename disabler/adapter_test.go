package disabler

import (
	"testing"

	casbin "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdapter(t *testing.T) {
	routes := []string{
		"GET:/api/user/{id}",
		"/api/user",
		"/user.User/GetUser",
		"",
	}

	a := NewAdapter(routes)
	assert.NotNil(t, a)

	m, err := model.NewModelFromString(DefaultModelText)
	require.NoError(t, err)

	e, err := casbin.NewEnforcer(m, a)
	require.NoError(t, err)

	ok1, _ := e.Enforce("/api/user", "GET")
	assert.True(t, ok1)

	ok2, _ := e.Enforce("/api/user", "POST")
	assert.True(t, ok2)

	ok3, _ := e.Enforce("/api/user/1", "GET")
	assert.True(t, ok3)

	ok4, _ := e.Enforce("/api/user/1", "POST")
	assert.False(t, ok4)

	ok5, _ := e.Enforce("/user.User/GetUser", "*")
	assert.True(t, ok5)
}
