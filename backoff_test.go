package retry_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.coder.com/retry"
)

func TestBackoff(t *testing.T) {
	t.Parallel()

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

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
				return
			}
		}

		t.Errorf("succeeded: took: %v", time.Since(start))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		start := time.Now()

		b := &retry.Backoff{
			Floor: time.Millisecond,
			Ceil:  time.Second * 5,
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*2)
		defer cancel()

		for time.Since(start) < time.Second {
			err := b.Wait(ctx)
			require.NoError(t, err, "took: %v", time.Since(start))
		}
	})
}
