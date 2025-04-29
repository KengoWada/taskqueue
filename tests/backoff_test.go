package taskqueue_test

import (
	"testing"
	"time"

	"github.com/KengoWada/taskqueue"
	"github.com/stretchr/testify/assert"
)

func TestBackoffPolicy(t *testing.T) {
	tests := []struct {
		name      string
		retries   uint
		maxDelay  time.Duration
		baseDelay time.Duration
	}{
		{
			name:      "first retry",
			retries:   1,
			baseDelay: 1 * time.Millisecond,
			maxDelay:  10 * time.Millisecond,
		},
		{
			name:      "second retry",
			retries:   1,
			baseDelay: 5 * time.Millisecond,
			maxDelay:  20 * time.Millisecond,
		},
		{
			name:      "third retry",
			retries:   3,
			baseDelay: 5 * time.Millisecond,
			maxDelay:  17 * time.Millisecond,
		},
		{
			name:      "fourth retry",
			retries:   4,
			baseDelay: 3 * time.Millisecond,
			maxDelay:  9 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &taskqueue.BackoffPolicy{BaseDelay: tt.baseDelay, MaxDelay: tt.maxDelay}
			delay := bp.Calculate(tt.retries)
			maxDelay := min(bp.BaseDelay*(1<<tt.retries), bp.MaxDelay)

			assert.GreaterOrEqual(t, delay, time.Duration(0))
			assert.Less(t, delay, maxDelay)
		})
	}
}
