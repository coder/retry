package retry

import (
	"context"
	"testing"
	"time"
)

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := New(time.Hour, time.Hour)
	for r.Wait(ctx) {
		t.Fatalf("attempt allowed even though context cancelled")
	}
}

func TestFirstTryImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := New(time.Hour, time.Hour)
	tt := time.Now()
	if !r.Wait(ctx) {
		t.Fatalf("attempt not allowed")
	}
	if time.Since(tt) > time.Second {
		t.Fatalf("attempt took too long")
	}
}

func TestReset(t *testing.T) {
	r := New(time.Hour, time.Hour)
	// Should be immediate
	ctx := context.Background()
	r.Wait(ctx)
	r.Reset()
	r.Wait(ctx)
}
