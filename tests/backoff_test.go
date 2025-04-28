package taskqueue_test

import (
	"testing"
	"time"

	"github.com/KengoWada/taskqueue"
	"github.com/stretchr/testify/assert"
)

func TestBackoffPolicy(t *testing.T) {
	tests := []struct {
		name          string
		backoffPolicy *taskqueue.BackoffPolicy
		retries       int
		expectedDelay time.Duration
		allowJitter   bool
	}{
		{
			name: "retry one",
			backoffPolicy: &taskqueue.BackoffPolicy{
				BaseDelay: 1 * time.Millisecond,
				MaxDelay:  10 * time.Millisecond,
			},
			retries:       1,
			expectedDelay: 1 * time.Millisecond,
		},
		{
			name: "retry three",
			backoffPolicy: &taskqueue.BackoffPolicy{
				BaseDelay: 1 * time.Millisecond,
				MaxDelay:  10 * time.Millisecond,
			},
			retries:       3,
			expectedDelay: 4 * time.Millisecond,
		},
		{
			name: "max delay",
			backoffPolicy: &taskqueue.BackoffPolicy{
				BaseDelay: 5 * time.Millisecond,
				MaxDelay:  10 * time.Millisecond,
			},
			retries:       5,
			expectedDelay: 10 * time.Millisecond, // should not exceed MaxDelay
		},
		{
			name: "jitter applied",
			backoffPolicy: &taskqueue.BackoffPolicy{
				BaseDelay:     1 * time.Millisecond,
				MaxDelay:      10 * time.Millisecond,
				UseJitter:     true,
				JitterRangeMs: 5,
			},
			retries:       1,
			expectedDelay: 1 * time.Millisecond, // with jitter this will vary slightly
			allowJitter:   true,
		},
		{
			name: "no jitter applied",
			backoffPolicy: &taskqueue.BackoffPolicy{
				BaseDelay: 1 * time.Millisecond,
				MaxDelay:  10 * time.Millisecond,
				UseJitter: false,
			},
			retries:       1,
			expectedDelay: 1 * time.Millisecond,
			allowJitter:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := tt.backoffPolicy.Calculate(tt.retries)

			if tt.allowJitter {
				assert.GreaterOrEqual(t, delay, tt.expectedDelay)
				assert.LessOrEqual(t, delay, tt.expectedDelay+5*time.Millisecond)
			} else {
				assert.Equal(t, delay, tt.expectedDelay)
			}
		})
	}
}
