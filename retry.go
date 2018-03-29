// Package retry contains utilities for retrying an action until it succeeds.
package retry

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

// Retry holds state about a retryable operation.
// Callers should create this via New.
type Retry struct {
	sleepDur func() time.Duration

	// preConditions are ran before each call to fn.
	preConditions []func() bool

	// postConditions are ran after each call to fn.
	postConditions []Condition
}

// New creates a new retry.
// The default retry will run forever, sleeping sleep.
func New(sleep time.Duration) *Retry {
	r := &Retry{
		sleepDur: func() time.Duration {
			return sleep
		},
	}

	r.appendPostConditions(func(err error) bool {
		return err != nil
	})

	return r
}

func (r *Retry) appendPreCondition(fn func() bool) {
	r.preConditions = append(r.preConditions, fn)
}
func (r *Retry) appendPostConditions(fns ...Condition) {
	for _, fn := range fns {
		r.postConditions = append(r.postConditions, fn)
	}
}

// Condition is a function that decides based on the given error whether to retry.
type Condition func(error) bool

// OnErrors returns a condition which retries on one of the provided errors.
func OnErrors(errs ...error) Condition {
	return func(err error) bool {
		for _, checkErr := range errs {
			if err == checkErr {
				return true
			}
		}
		return false
	}
}

// NotOnErrors returns a condition which retries only if the error
// does not match one of the provided errors.
func NotOnErrors(errs ...error) Condition {
	return func(err error) bool {
		for _, checkErr := range errs {
			if err == checkErr {
				return false
			}
		}
		return true
	}
}

// Condition appends the passed retry conditions.
// All conditions must return true for the retry to progress.
// The error passed to the retry conditions will be the result
// of errors.Cause() on the original error from  the run function.
func (r *Retry) Conditions(fns ...Condition) *Retry {
	r.appendPostConditions(fns...)
	return r
}

// Condition appends the passed retry condition.
// The condition must return true for the retry to progress.
// Deprecated: Use Conditions instead.
func (r *Retry) Condition(fn Condition) *Retry {
	return r.Conditions(fn)
}

func (r *Retry) preCheck() bool {
	for _, fn := range r.preConditions {
		if !fn() {
			return false
		}
	}
	return true
}

func (r *Retry) postCheck(err error) bool {
	err = errors.Cause(err)
	for _, fn := range r.postConditions {
		if !fn(err) {
			return false
		}
	}
	return true
}

// Attempts sets the maximum amount of retry attempts
// before the current error is returned.
func (r *Retry) Attempts(n int) *Retry {
	var iterations int
	r.appendPreCondition(func() bool {
		ok := iterations < n
		iterations++
		return ok
	})
	return r
}

// Context bounds the retry to when the context expires.
func (r *Retry) Context(ctx context.Context) *Retry {
	r.appendPreCondition(func() bool {
		return ctx.Err() == nil
	})
	return r
}

// Backoff turns retry into an exponential backoff
// with a maximum sleep of ceil.
func (r *Retry) Backoff(ceil time.Duration) *Retry {
	const growth = 2

	// Start delay at half so that
	// the first iteration of sleepDur doubles it.
	delay := r.sleepDur() / growth

	if delay == 0 {
		panic("retry: delay must not be zero (is it less than 2 nanoseconds?) ")
	}

	r.sleepDur = func() time.Duration {
		if delay < ceil {
			delay = delay * growth
			if delay > ceil {
				delay = ceil
			}
		}
		return delay
	}

	return r
}

// Timeout returns the retry with a bounding timeout.
// If the passed timeout is 0, Timeout does nothing. This has been done
// to match the behaviour of the previous retry API and to make it easy
// for functions that call into retry to offer optional timeouts.
func (r *Retry) Timeout(to time.Duration) *Retry {
	if to == 0 {
		return r
	}

	deadline := time.Now().Add(to)

	r.appendPreCondition(func() bool {
		return time.Now().Before(deadline)
	})

	return r
}

// Jitter adds some random jitter to the retry's sleep.
//
// Ratio must be between 0 and 1, and determines how jittery
// the sleeps will be. For example, a rat of 0.1 and a sleep of 1s restricts the
// jitter to the range of 900ms to 1.1 seconds.
func (r *Retry) Jitter(rat float64) *Retry {
	if !(rat < 1 && rat > 0) {
		panic("retry: rat must be (0, 1)")
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	underlyingSleep := r.sleepDur
	r.sleepDur = func() time.Duration {
		dur := underlyingSleep()

		var (
			minDuration = float64(dur) * (1 - rat)
			maxDuration = float64(dur) * (1 + rat)
		)

		dur = time.Duration(minDuration) + time.Duration(rnd.Int63n(int64(maxDuration)-int64(minDuration)))
		return dur
	}

	return r
}

// Log adds a function to log any returned errors.
// It is added as a post condition that always returns true.
// If you want an error to stop the retry and not be logged,
// use Log() after the Condition.
// If you want an error to stop the retry and be logged,
// use Log() before the Condition.
func (r *Retry) Log(logFn func(error)) *Retry {
	return r.Conditions(func(err error) bool {
		logFn(err)
		return true
	})
}

// Run runs the retry.
// The retry must not be ran twice.
func (r *Retry) Run(fn func() error) error {
	err := errors.New("didn't run a single iteration?")
	for r.preCheck() {
		err = fn()
		if !r.postCheck(err) {
			return err
		}
		time.Sleep(r.sleepDur())
	}
	return err
}
