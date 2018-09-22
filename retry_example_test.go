package retry_test

import (
	"context"
	"log"
	"net"
	"time"

	"go.coder.com/retry"
)

func ExampleBackoffSuccess() {
	start := time.Now()

	b := &retry.Backoff{
		Floor: time.Millisecond,
		Ceil:  time.Second * 5,
	}

	ctx := context.Background()

	for time.Since(start) < time.Second {
		err := b.Wait(ctx)
		if err != nil {
			log.Fatalf("failed: took: %v: err: %v", time.Since(start), err)
		}
	}

	log.Printf("success: took: %v", time.Since(start))
}

func ExampleBackoffError() {
	start := time.Now()

	b := &retry.Backoff{
		Floor: time.Millisecond,
		Ceil:  time.Second * 5,
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()

	for time.Since(start) < time.Second {
		err := b.Wait(ctx)
		if err != nil {
			log.Fatalf("failed: took: %v: err: %v", time.Since(start), err)
		}
	}

	log.Printf("success: took: %v", time.Since(start))
}

func ExampleListener() {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer l.Close()

	l = retry.Listener{
		Listener: l,
	}

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalf("failed to accept: %v", err)
		}
		defer c.Close()

		// ...
	}
}
