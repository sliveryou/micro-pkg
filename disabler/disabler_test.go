package disabler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuncDisabler_AllowAPI(t *testing.T) {
	fd, err := NewFuncDisabler(Config{
		DisabledAPIs: []string{
			"GET:/api/file",
			"/api/user",
			"POST:/api/file/upload",
			"/api/auth/{user_id}",
		},
	})
	require.NoError(t, err)

	assert.False(t, fd.AllowAPI("GET", "/api/file"))
	assert.False(t, fd.AllowAPI("POST", "/api/file/upload"))
	assert.True(t, fd.AllowAPI("GET", "/api/file/list"))
	assert.False(t, fd.AllowAPI("GET", "/api/user"))
	assert.False(t, fd.AllowAPI("POST", "/api/user"))
	assert.True(t, fd.AllowAPI("GET", "/api/user/list"))
	assert.True(t, fd.AllowAPI("GET", "/api/file/upload"))
	assert.False(t, fd.AllowAPI("POST", "/api/file/upload"))
	assert.False(t, fd.AllowAPI("POST", "/api/auth/1"))
	assert.False(t, fd.AllowAPI("GET", "/api/auth/2?name=xxx"))
	assert.True(t, fd.AllowAPI("GET", "/api/auth"))
	assert.True(t, fd.AllowAPI("GET", "/api/auth/1/info"))
}

func TestFuncDisabler_AllowRPC(t *testing.T) {
	fd, err := NewFuncDisabler(Config{
		DisabledRPCs: []string{
			"/auth.Auth/*",
			"/contract.Contract/GetPasses",
			"/file.File/*",
			"/pay.Pay/GetPlan",
		},
	})
	require.NoError(t, err)

	assert.False(t, fd.AllowRPC("/auth.Auth/GetPersonalAuth"))
	assert.False(t, fd.AllowRPC("/contract.Contract/GetPasses"))
	assert.True(t, fd.AllowRPC("/contract.Contract/GetPass"))
	assert.False(t, fd.AllowRPC("/file.File/GetFiles"))
	assert.False(t, fd.AllowRPC("/file.File/UploadFile"))
	assert.False(t, fd.AllowRPC("/pay.Pay/GetPlan"))
	assert.True(t, fd.AllowRPC("/pay.Pay/GetPlans"))
}
