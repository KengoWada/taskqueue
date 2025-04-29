package taskqueue

import (
	"math/rand"
	"time"
)

var DefaultBackoffPolicy = BackoffPolicy{
	BaseDelay:     1 * time.Second,
	MaxDelay:      30 * time.Second,
	UseJitter:     true,
	JitterRangeMs: 300,
}

// Backoff defines the interface for calculating backoff delays between retries.
// Implementations of this interface can provide custom logic for determining
// how long to wait before retrying a failed task based on the number of retries.
type Backoff interface {
	// Calculate returns the duration to wait before the next retry attempt.
	// The input parameter retries indicates how many times the task has already been retried.
	Calculate(retries uint) time.Duration
}

// BackoffPolicy defines the configuration for handling retries with backoff logic.
// It provides settings for base delay, maximum delay, jitter, and the range for jitter.
type BackoffPolicy struct {
	BaseDelay     time.Duration // Base delay before the first retry (e.g., 1 second)
	MaxDelay      time.Duration // Maximum delay before retrying again (e.g., 60 seconds)
	UseJitter     bool          // Whether to add random jitter to the delay
	JitterRangeMs int           // Maximum jitter in milliseconds (e.g., 500ms)
}

// Calculate calculates the delay before the next retry based on the backoff policy.
// It considers the retry count, base delay, maximum delay, and jitter.
func (b *BackoffPolicy) Calculate(retries uint) time.Duration {
	// Calculate the exponential backoff delay, doubling with each retry.
	delay := min(b.BaseDelay*time.Duration(1<<uint(retries-1)), b.MaxDelay)

	if b.UseJitter && b.JitterRangeMs > 0 {
		jitter := time.Duration(rand.Intn(b.JitterRangeMs)) * time.Millisecond
		delay += jitter
	}

	return delay
}
