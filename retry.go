// Package retry contains utilities for retrying an action until it succeeds.
package retry

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"
)

// Attempts calls f attempts times until it doesn't return an error.
func Attempts(attempts int, delay time.Duration, f func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = f(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return err
}

// Timeout calls f until timeout is exceeded.
func Timeout(timeout, delay time.Duration, f func() error) error {
	var err error
	for maxTime := time.Now().Add(timeout); time.Now().Before(maxTime); time.Sleep(delay) {
		if err = f(); err == nil {
			return nil
		}
	}
	return err
}

var errCeilLessThanFloor = errors.New("ceiling cannot be less than the floor")

// Backoff implements an exponential backoff algorithm.
// It calls f before timeout is exceeded using ceil as a maximum sleep
// interval and floor as the start interval.
// Since calls to f may take an undefined amount of time, Backoff cannot guarantee
// it will return by timeout.
// If timeout is 0, it will run until the function returns a nil error.
func Backoff(timeout, ceil, floor time.Duration, f func() error) error {
	var ctx context.Context
	if timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	return BackoffContext(ctx, ceil, floor, f)
}

// BackoffWhile implements an exponential backoff algorithm.
//
// It calls f before timeout is exceeded, using ceil as a maximum sleep interval,
// and floor as the starting interval.
//
// It will continue calling f until either the timeout is exceeded or cond(f()) == false.
//
// If timeout is 0, f will be called until cond(f()) == false.
func BackoffWhile(timeout, ceil, floor time.Duration, f func() error, cond func(error) bool) error {
	var ctx context.Context
	if timeout != 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	return backoffContextWhile(ctx, ceil, floor, f, cond)
}

// BackoffContext implements an exponential backoff algorithm.
// It calls f before the context is cancelled using ceil as a maximum sleep
// interval and floor as the start interval.
func BackoffContext(ctx context.Context, ceil, floor time.Duration, f func() error) error {
	return backoffContextWhile(ctx, ceil, floor, f, func(err error) bool {
		if err != nil {
			return true
		}
		return false
	})
}

func backoffContextWhile(ctx context.Context, ceil, floor time.Duration, f func() error, cond func(error) bool) error {
	if ceil < floor {
		return errCeilLessThanFloor
	}

	delay := floor
	var err error

	for {
		err = f()
		if !cond(err) {
			return err
		}

		t := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			t.Stop()
			return err
		case <-t.C:
		}

		if delay < ceil {
			delay = delay * 2
			if delay > ceil {
				delay = ceil
			}
		}
	}
}

type Listener struct {
	LogTmpErr func(err error)
	net.Listener
}

const (
	initialListenerDelay = 5 * time.Millisecond
	maxListenerDelay     = time.Second
)

func (l *Listener) Accept() (net.Conn, error) {
	var retryDelay time.Duration
	for {
		c, err := l.Listener.Accept()
		if err != nil {
			ne, ok := err.(net.Error)
			if ok && ne.Temporary() {
				if retryDelay == 0 {
					retryDelay = initialListenerDelay
				} else {
					retryDelay *= 2
					if retryDelay > maxListenerDelay {
						retryDelay = maxListenerDelay
					}
				}
				if l.LogTmpErr == nil {
					log.Printf("retry: temp error accepting next connection: %v; retrying in %v", err, retryDelay)
				} else {
					l.LogTmpErr(err)
				}
				time.Sleep(retryDelay)
				continue
			}
			return nil, err
		}
		return c, nil
	}
}
