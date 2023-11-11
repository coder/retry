package main

import (
	"context"
	"fmt"
	"time"

	"github.com/coder/retry"
)

func main() {
	r := retry.New(time.Second, time.Second*10)

	ctx := context.Background()

	last := time.Now()
	for r.Wait(ctx) {
		// Do something that might fail
		fmt.Printf("%v: hi\n", time.Since(last).Round(time.Second))
		last = time.Now()
	}
}
