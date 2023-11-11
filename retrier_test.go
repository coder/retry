package retry

import (
	"context"
	"math"
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

func TestScalesExponentially(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := New(time.Second, time.Second*10)
	r.Rate = 2

	start := time.Now()

	for i := 0; i < 3; i++ {
		t.Logf("delay: %v", r.Delay)
		r.Wait(ctx)
		t.Logf("sinceStart: %v", time.Since(start).Round(time.Second))
	}

	sinceStart := time.Since(start).Round(time.Second)
	if sinceStart != time.Second*6 {
		t.Fatalf("did not scale correctly: %v", sinceStart)
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

func TestJitter_Normal(t *testing.T) {
	t.Parallel()

	r := New(time.Millisecond, time.Millisecond)
	r.Jitter = 0.5

	var (
		sum   time.Duration
		waits []float64
		ctx   = context.Background()
	)
	for i := 0; i < 1000; i++ {
		start := time.Now()
		r.Wait(ctx)
		took := time.Since(start)
		waits = append(waits, (took.Seconds() * 1000))
		sum += took
	}

	avg := float64(sum) / float64(len(waits))
	std := stdDev(waits)
	if std > avg*0.1 {
		t.Fatalf("standard deviation too high: %v", std)
	}

	t.Logf("average: %v", time.Duration(avg))
	t.Logf("std dev: %v", std)
	t.Logf("sample: %v", waits[len(waits)-10:])
}

// stdDev returns the standard deviation of the sample.
func stdDev(sample []float64) float64 {
	if len(sample) == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range sample {
		mean += v
	}
	mean /= float64(len(sample))

	variance := 0.0
	for _, v := range sample {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(sample))

	return math.Sqrt(variance)
}
