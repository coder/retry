package retry

import (
	"context"
	"time"
)

// Retrier represents a retry instance.
// Use New instead of creating this object directly.
type Retrier struct {
	delay       time.Duration
	floor, ceil time.Duration
	err         *error
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
	if r.err != nil && *r.err == nil {
		// We've succeeded!
		return false
	}
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

func (r *Retrier) SetError(err error) {
	r.err = &err
}

func (r *Retrier) Error() error {
	if r.err == nil {
		return nil
	}
	return *r.err
}
