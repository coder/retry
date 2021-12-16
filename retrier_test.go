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
