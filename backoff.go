package taskqueue

import (
	"math/rand"
	"time"
)

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
func (b *BackoffPolicy) Calculate(retries int) time.Duration {
	// Calculate the exponential backoff delay, doubling with each retry.
	delay := min(b.BaseDelay*time.Duration(1<<uint(retries-1)), b.MaxDelay)

	if b.UseJitter && b.JitterRangeMs > 0 {
		jitter := time.Duration(rand.Intn(b.JitterRangeMs)) * time.Millisecond
		delay += jitter
	}

	return delay
}
