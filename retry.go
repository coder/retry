// Package retry contains utilities for retrying an action until it succeeds.
package retry

import (
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
