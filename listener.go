package retry

import (
	"context"
	"log"
	"net"
	"time"
)

type Listener struct {
	LogTmpErr func(err error)
	net.Listener
}

func (l Listener) Accept() (net.Conn, error) {
	b := &Backoff{
		Floor: 5 * time.Millisecond,
		Ceil:  time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for {
		c, err := l.Listener.Accept()
		if err == nil {
			return c, nil
		}

		ne, ok := err.(net.Error)
		if !ok || !ne.Temporary() {
			return nil, err
		}

		if l.LogTmpErr == nil {
			log.Printf("retry: temp error accepting next connection: %v", err)
		} else {
			l.LogTmpErr(err)
		}

		err = b.Wait(ctx)
		if err != nil {
			return nil, err
		}
	}
}
