// Package retry contains utilities for retrying an action until it succeeds.
package retry

import (
	"context"
	"log"
	"net"
	"time"
)

// Attempts calls f attempts times until it doesn't return an error.
func Attempts(delay time.Duration, attempts int, f func() error) error {
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
func Timeout(delay time.Duration, timeout time.Duration, f func() error) error {
	var err error
	for maxTime := time.Now().Add(timeout); time.Now().Before(maxTime); time.Sleep(delay) {
		if err = f(); err == nil {
			return nil
		}
	}
	return err
}

// Backoff implements an exponential backoff algorithm.
// It calls f before timeout is exceeded using ceil as a maximum sleep
// interval and floor as the start interval.
// Since calls to f may take an undefined amount of time, Backoff cannot guarantee
// it will return by timeout.
// If timeout is 0, it will run until the function returns a nil error.
func Backoff(timeout time.Duration, ceil time.Duration, floor time.Duration, f func() error) error {
	var deadline time.Time
	if timeout == 0 {
		deadline = time.Now().AddDate(1337, 0, 0)
	} else {
		deadline = time.Now().Add(timeout)
	}

	delay := floor
	var err error

	for {
		err = f()
		if err == nil {
			return nil
		}
		// since we're about to sleep for delay, we may as well
		// factor it into our deadline calculation.
		if time.Now().Add(delay).After(deadline) {
			return err
		}
		time.Sleep(delay)
		if delay < ceil {
			delay = delay * 2
			if delay > ceil {
				delay = ceil
			}
		}
	}
}

// BackoffContext implements an exponential backoff algorithm.
// It calls f before the context is cancelled using ceil as a maximum sleep
// interval and floor as the start interval.
func BackoffContext(ctx context.Context, ceil time.Duration, floor time.Duration, f func() error) error {
	delay := floor
	var err error

	for {
		err = f()
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return err
		default:
		}

		time.Sleep(delay)
		if delay < ceil {
			delay = delay * 2
			if delay > ceil {
				delay = ceil
			}
		}
	}
}

type Listener struct {
	LogTmpErr func(err error, retryDelay time.Duration)
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
					l.LogTmpErr(err, retryDelay)
				}
				time.Sleep(retryDelay)
				continue
			}
			return nil, err
		}
		return c, nil
	}
}
