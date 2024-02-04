package server

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/health/checker/compositechecker"
)

func TestNewHealthServer(t *testing.T) {
	cc := compositechecker.NewChecker()
	_, err := NewHealthServer("test.rpc", cc)
	require.NoError(t, err)

	_, err = NewHealthServer("", nil)
	require.EqualError(t, err, "health: illegal health configure")
}
