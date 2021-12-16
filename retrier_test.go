package retry

import (
	"context"
	"fmt"
	"io"
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

func TestError1(t *testing.T) {
	r := New(time.Millisecond, time.Millisecond*10)
	var n int
	for ; r.Wait(context.Background()); n++ {
		if n < 9 {
			r.SetError(fmt.Errorf("n is too small"))
			continue
		}
		r.SetError(nil)
	}
	if n != 10 {
		t.Fatalf("expected n == 10, but n == %v", n)
	}
}

func TestError2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	r := New(time.Millisecond, time.Millisecond*10)
	var n int
	for ; r.Wait(ctx); n++ {
		r.SetError(io.EOF)
	}
	if r.Error() != io.EOF {
		t.Fatalf("expected error %v but got %v", io.EOF, r.Error())
	}
}
