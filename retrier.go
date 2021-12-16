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
		delay: floor,
		floor: floor,
		ceil:  ceil,
	}
}

func (r *Retrier) Wait(ctx context.Context) bool {
	const growth = 2
	r.delay *= growth
	if r.delay > r.ceil {
		r.delay = r.ceil
	}
	select {
	case <-time.After(r.delay):
		return true
	case <-ctx.Done():
		return false
	}
}
