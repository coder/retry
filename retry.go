// Package retry contains utilities for retrying an action until it succeeds.
package retry

import (
	"net"
	"time"

	"go.uber.org/zap"
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
	MaxDelay time.Duration
	Logger   zap.Logger
	net.Listener
}

func (l *Listener) Accept() (net.Conn, error) {
	var delay time.Duration
	for {
		c, err := l.Listener.Accept()
		if err != nil {
			ne, ok := err.(net.Error)
			if ok && ne.Temporary() {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
					if delay > time.Second {
						delay = time.Second
					}
				}
				l.Logger.Error("failed to accept next connection",
					zap.Error(err),
					zap.Duration("retrying", delay),
				)
				time.Sleep(delay)
				continue
			}
			return nil, err
		}
		return c, nil
	}
}
