package taskqueue

import (
	"math/rand"
	"time"
)

var DefaultBackoffPolicy = BackoffPolicy{
	BaseDelay: 1 * time.Second,
	MaxDelay:  30 * time.Second,
}

// Backoff defines the interface for calculating backoff delays between retries.
// Implementations of this interface can provide custom logic for determining
// how long to wait before retrying a failed task based on the number of retries.
type Backoff interface {
	// Calculate returns the duration to wait before the next retry attempt.
	// The input parameter retries indicates how many times the task has already been retried.
	Calculate(retries uint) time.Duration
}

// BackoffPolicy specifies the parameters for calculating exponential backoff with jitter.
// It defines the base delay and the maximum delay allowed between retries.
//
// Fields:
//   - BaseDelay: Specifies the initial delay duration.
//   - MaxDelay: Sets the upper bound for the backoff delay, preventing unbounded wait times.
type BackoffPolicy struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// Calculate returns a jittered backoff duration based on the number of retries.
// It uses exponential backoff with full jitter, where the delay is randomly chosen
// between 0 and the calculated maximum delay.
// The maximum delay grows exponentially with each retry, capped by MaxDelay.
//
// Example:
//
//	BaseDelay = 100ms, MaxDelay = 3s, retries = 3
//	Max calculated delay = min(100ms * 2^3, 3s) = 800ms
//	Returned delay = random duration in [0, 800ms)
func (b *BackoffPolicy) Calculate(retries uint) time.Duration {
	maxDelay := min(b.BaseDelay*(1<<retries), b.MaxDelay)
	jitter := rand.Int63n(int64(maxDelay))
	return time.Duration(jitter)
}
