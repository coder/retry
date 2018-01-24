// Package retry contains utilities for retrying an action until it succeeds.
package retry

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// Retry contains
type Retry struct {
	fn       func() error
	sleepDur func() time.Duration

	cond func(err error) bool

	// conts contains a list of functions which must return true in order
	// for the retry to continue.
	conts []func() bool

	iteration int
}

// New creates a new retry.
// The default retry will run forever, sleeping sleep.
func New(fn func() error, sleep time.Duration) *Retry {
	return &Retry{
		fn: fn,
		sleepDur: func() time.Duration {
			return sleep
		},
		cond: func(err error) bool {
			return true
		},
	}
}

// Cond lets the caller override the default condition behaviour.
// The retry only continues if fn returns true
func (r *Retry) Cond(fn func(err error) bool) *Retry {
	r.cond = fn
	return r
}

// cont iterates over each continue function, returning
// false if any of them do.
func (r *Retry) cont() bool {
	for _, fn := range r.conts {
		if !fn() {
			return false
		}
	}
	return true
}

func (r *Retry) appendCont(fn func() bool) {
	r.conts = append(r.conts, fn)
}

// Attempts sets the maximum amount of retry attempts
// before the current error is returned.
func (r *Retry) Attempts(n int) *Retry {
	r.appendCont(func() bool {
		return r.iteration < n
	})
	return r
}

// Context bounds the retry to when the context expires.
func (r *Retry) Context(ctx context.Context) *Retry {
	r.appendCont(func() bool {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	})
	return r
}

// Backoff turns retry into an exponential backoff
// with a maximum sleep of ceil.
func (r *Retry) Backoff(ceil time.Duration) *Retry {
	const growth = 2

	// start delay at half so that
	// the first iteration of sleepDur doubles it.
	delay := r.sleepDur() / growth

	if delay == 0 {
		panic("retry: delay must not be zero (is it less than 2 nanoseconds?) ")
	}

	r.sleepDur = func() time.Duration {
		if delay < ceil {
			delay = delay * 2
			if delay > ceil {
				delay = ceil
			}
		}
		return delay
	}

	return r
}

// Timeout returns the retry with a bounding timeout.
func (r *Retry) Timeout(to time.Duration) *Retry {
	deadline := time.Now().Add(to)

	r.appendCont(func() bool {
		return time.Now().Before(deadline)
	})

	return r
}

// Run runs the retry.
// The retry must not be ran twice.
func (r *Retry) Run() error {
	err := errors.Errorf("didn't run a single iteration?")
	for ; r.cont(); r.iteration++ {
		err = r.fn()
		if !r.cond(err) || err == nil {
			return err
		}
		time.Sleep(r.sleepDur())
	}
	return err
}
