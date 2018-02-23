package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/codercom/retry"
)

func main() {
	start := time.Now()

	err := retry.New(time.Millisecond).
		Backoff(time.Millisecond * 50).
		Timeout(time.Second * 2).
		Run(
			func() error {
				if time.Since(start) < time.Second {
					return errors.New("not enough time has elapsed")
				}
				return nil
			},
		)
	fmt.Printf("err: %v, took %v\n", err, time.Since(start))
}
