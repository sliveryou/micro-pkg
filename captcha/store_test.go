package captcha

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/go-tool/v2/randx"
)

func TestNewStore(t *testing.T) {
	assert.NotPanics(t, func() {
		store := NewStore(getStore(), "", 300)
		assert.NotNil(t, store)
	})
}

func TestStore_Set(t *testing.T) {
	id := randx.NewString(12)
	value := randx.NewNumber(8)
	store := NewStore(getStore(), "", 300)

	err := store.Set(id, value)
	require.NoError(t, err)

	get, err := store.kvStore.Get(id)
	require.NoError(t, err)

	assert.Equal(t, get, value)
}

func TestStore_Get(t *testing.T) {
	id := randx.NewString(12)
	value := randx.NewNumber(8)
	store := NewStore(getStore(), "", 300)

	err := store.Set(id, value)
	require.NoError(t, err)

	get := store.Get(id, true)
	assert.Equal(t, get, value)

	get = store.Get(id, false)
	assert.Empty(t, get)
}

func TestStore_Verify(t *testing.T) {
	id := randx.NewString(12)
	value := randx.NewNumber(8)
	store := NewStore(getStore(), "", 300)

	err := store.Set(id, value)
	require.NoError(t, err)

	ok := store.Verify(id, "error answer", false)
	assert.False(t, ok)

	ok = store.Verify(id, value, true)
	assert.True(t, ok)

	get := store.Get(id, false)
	assert.Empty(t, get)
}
