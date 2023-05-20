package retry

import (
	"context"
	"time"
)

// Retrier implements an exponentially backing off retry instance.
// Use New instead of creating this object directly.
type Retrier struct {
	delay       time.Duration
	floor, ceil time.Duration
}

// New creates a retrier that exponentially backs off from floor to ceil pauses.
func New(floor, ceil time.Duration) *Retrier {
	return &Retrier{
		delay: 0,
		floor: floor,
		ceil:  ceil,
	}
}

// Next returns the next delay duration without modifying the retry state.
// This is useful for logging.
func (r *Retrier) Next() time.Duration {
	const growth = 2
	delay := r.delay * growth
	if delay > r.ceil {
		delay = r.ceil
	}
	return delay
}

// Wait waits for the next retry and returns true if the retry should be attempted.
func (r *Retrier) Wait(ctx context.Context) bool {
	r.delay = r.Next()
	select {
	case <-time.After(r.delay):
		if r.delay < r.floor {
			r.delay = r.floor
		}
		return true
	case <-ctx.Done():
		return false
	}
}

// Reset resets the retrier to its initial state.
func (r *Retrier) Reset() {
	r.delay = 0
}
