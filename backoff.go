package retry

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// Backoff holds state about a backoff loop in which
// there should be a delay in iterations.
type Backoff struct {
	// These two fields must be initialized.
	// Floor should never be greater than or equal
	// to the Ceil in general. If it is, the Backoff
	// will stop backing off and just sleep for the Floor
	// in Wait().
	Floor time.Duration
	Ceil  time.Duration

	delay time.Duration
}

func (b *Backoff) backoff() {
	if b.Floor >= b.Ceil {
		return
	}

	const growth = 2
	b.delay *= growth
	if b.delay > b.Ceil {
		b.delay = b.Ceil
	}
}

// Wait should be called at the end of the loop. It will sleep
// for the necessary duration before the next iteration of the loop
// can begin.
// If the context is cancelled, Wait will return early with a non-nil error.
func (b *Backoff) Wait(ctx context.Context) error {
	if b.delay < b.Floor {
		b.delay = b.Floor
	}

	select {
	case <-ctx.Done():
		return errors.Wrapf(ctx.Err(), "failed to sleep delay %v for retry attempt", b.delay)
	case <-time.After(b.delay):
	}

	b.backoff()

	return nil
}
