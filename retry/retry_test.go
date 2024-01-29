package retry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	action := func(attempt uint) error {
		return nil
	}

	err := Retry(action)
	require.NoError(t, err)
}

func TestRetryUntilNoErrorReturned(t *testing.T) {
	const errorUntilAttemptNumber = uint(5)

	var attemptsMade uint

	action := func(attempt uint) error {
		attemptsMade = attempt

		if attempt == errorUntilAttemptNumber {
			return nil
		}

		return errors.New("error")
	}

	err := Retry(action)
	require.NoError(t, err)
	assert.Equal(t, errorUntilAttemptNumber, attemptsMade)
}

func TestShouldAttempt(t *testing.T) {
	shouldAttempt := shouldAttempt(1)
	assert.True(t, shouldAttempt)
}

func TestShouldAttemptWithMultiStrategies(t *testing.T) {
	trueStrategy := func(attempt uint) bool {
		return true
	}

	falseStrategy := func(attempt uint) bool {
		return false
	}

	should := shouldAttempt(1, trueStrategy)
	assert.True(t, should)

	should = shouldAttempt(1, falseStrategy)
	assert.False(t, should)

	should = shouldAttempt(1, trueStrategy, trueStrategy, trueStrategy)
	assert.True(t, should)

	should = shouldAttempt(1, falseStrategy, falseStrategy, falseStrategy)
	assert.False(t, should)

	should = shouldAttempt(1, trueStrategy, trueStrategy, falseStrategy)
	assert.False(t, should)
}
