// Package retry contains utilities for retrying an action until it succeeds.
package retry

import (
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

type Listener struct {
	LogTmpErr func(err error, retry time.Duration)
	net.Listener
}

const (
	initialListenerDelay = 5 * time.Millisecond
	maxListenerDelay     = time.Second
)

func (l *Listener) Accept() (net.Conn, error) {
	var delay time.Duration
	for {
		c, err := l.Listener.Accept()
		if err != nil {
			ne, ok := err.(net.Error)
			if ok && ne.Temporary() {
				if delay == 0 {
					delay = initialListenerDelay
				} else {
					delay *= 2
					if delay > maxListenerDelay {
						delay = maxListenerDelay
					}
				}
				l.LogTmpErr(err, delay)
				time.Sleep(delay)
				continue
			}
			return nil, err
		}
		return c, nil
	}
}
