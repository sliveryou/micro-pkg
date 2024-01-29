package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	strategy := Limit(3)

	assert.True(t, strategy(0))
	assert.True(t, strategy(1))
	assert.True(t, strategy(2))
	assert.False(t, strategy(3))
	assert.False(t, strategy(4))
	assert.False(t, strategy(5))
}

func TestDelay(t *testing.T) {
	const delayDuration = 10 * time.Millisecond

	strategy := Delay(delayDuration)

	now := time.Now()
	assert.True(t, strategy(0) && float64(delayDuration)*0.95 <= float64(time.Since(now)))

	now = time.Now()
	assert.True(t, strategy(5) && (delayDuration/10) >= time.Since(now))
}

func TestWaitWithDuration(t *testing.T) {
	const waitDuration = 10 * time.Millisecond

	strategy := Wait(waitDuration)

	now := time.Now()
	assert.True(t, strategy(0) && 1*time.Millisecond >= time.Since(now))

	now = time.Now()
	assert.True(t, strategy(1) && float64(waitDuration)*0.95 <= float64(time.Since(now)))
}

func TestWaitWithMultiDurations(t *testing.T) {
	waitDurations := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
	}

	strategy := Wait(waitDurations...)

	now := time.Now()
	assert.True(t, strategy(0) && 1*time.Millisecond >= time.Since(now))

	now = time.Now()
	assert.True(t, strategy(1) && float64(waitDurations[0])*0.95 <= float64(time.Since(now)))

	now = time.Now()
	assert.True(t, strategy(3) && float64(waitDurations[2])*0.95 <= float64(time.Since(now)))

	now = time.Now()
	assert.True(t, strategy(999) && float64(waitDurations[len(waitDurations)-1])*0.95 <= float64(time.Since(now)))
}

func TestFail(t *testing.T) {
	type args struct {
		attemptLimit uint
		failAction   Action
		attempt      uint
	}
	failAction := func(attempt uint) error {
		t.Log("fail, attempt =", attempt)
		return nil
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1-2",
			args: args{
				attemptLimit: 1,
				failAction:   failAction,
				attempt:      2,
			},
			want: true,
		},
		{
			name: "2-0",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      0,
			},
			want: true,
		},
		{
			name: "2-1",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      1,
			},
			want: true,
		},
		{
			name: "2-2",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      2,
			},
			want: true,
		},
		{
			name: "2-3",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      3,
			},
			want: true,
		},
		{
			name: "2-4",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      4,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fail(tt.args.attemptLimit, tt.args.failAction)(tt.args.attempt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFailLimit(t *testing.T) {
	type args struct {
		attemptLimit uint
		failAction   Action
		attempt      uint
	}
	failAction := func(attempt uint) error {
		t.Log("fail, attempt =", attempt)
		return nil
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1-2",
			args: args{
				attemptLimit: 1,
				failAction:   failAction,
				attempt:      2,
			},
			want: false,
		},
		{
			name: "2-0",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      0,
			},
			want: true,
		},
		{
			name: "2-1",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      1,
			},
			want: true,
		},
		{
			name: "2-2",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      2,
			},
			want: false,
		},
		{
			name: "2-3",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      3,
			},
			want: false,
		},
		{
			name: "2-4",
			args: args{
				attemptLimit: 2,
				failAction:   failAction,
				attempt:      4,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FailLimit(tt.args.attemptLimit, tt.args.failAction)(tt.args.attempt)
			assert.Equal(t, tt.want, got)
		})
	}
}
