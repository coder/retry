package retry

import (
	"log"
	"net"
	"time"
)

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
