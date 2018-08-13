package retry

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	t.Parallel()

	t.Run("Attempts", func(t *testing.T) {
		t.Parallel()

		t.Run("Respects count and sleeps between attempts", func(t *testing.T) {
			t.Parallel()

			count := 0
			start := time.Now()

			const sleep = time.Millisecond * 10

			New(sleep).Attempts(5).Run(func() error {
				count++
				return errors.Errorf("asdfasdf")
			})

			assert.Equal(t, 5, count)
			assert.WithinDuration(t, start.Add(sleep*5), time.Now(), sleep)
		})

		t.Run("returns as soon as error is nil", func(t *testing.T) {
			t.Parallel()

			start := time.Now()

			var count int
			New(time.Minute).Attempts(100).Run(func() error {
				count++
				return nil
			})

			assert.Equal(t, 1, count)
			assert.WithinDuration(t, time.Now(), start, time.Millisecond)
		})
	})

	t.Run("Backoff", func(t *testing.T) {
		t.Parallel()

		t.Run("return when nil", func(t *testing.T) {
			t.Parallel()

			var count int
			err := New(time.Millisecond).Timeout(time.Minute).Backoff(time.Second).Run(func() error {
				count++
				if count == 10 {
					return nil
				}
				return io.EOF
			})

			assert.Equal(t, 10, count)
			assert.NoError(t, err)
		})
	})

	t.Run("Timeout", func(t *testing.T) {
		t.Parallel()

		t.Run("Respects timeout and sleeps between attempts", func(t *testing.T) {
			t.Parallel()
			count := 0
			start := time.Now()

			const sleep = time.Millisecond * 10

			var errSomething = errors.New("something")
			err := New(sleep).Timeout(sleep * 5).Run(func() error {
				count++
				return errSomething
			})

			assert.Equal(t, 5, count)
			assert.WithinDuration(t, start.Add(sleep*5), time.Now(), sleep)
			assert.Equal(t, errSomething, errors.Cause(err))
		})

		t.Run("returns as soon as error is nil", func(t *testing.T) {
			t.Parallel()
			start := time.Now()

			New(time.Minute).Timeout(time.Hour).Run(func() error {
				return nil
			})

			assert.WithinDuration(t, time.Now(), start, time.Millisecond*10)
		})
	})

	t.Run("Cond", func(t *testing.T) {
		t.Parallel()
		count := 0

		err := New(time.Millisecond).
			Conditions(
				NotOnErrors(io.ErrShortWrite),
				OnErrors(io.ErrUnexpectedEOF, io.EOF, io.ErrShortWrite),
			).
			Run(func() error {
				count++
				if count == 5 {
					return io.ErrShortWrite
				}
				if count%2 == 0 {
					return errors.WithStack(io.ErrUnexpectedEOF)
				}
				return io.EOF
			})

		assert.Equal(t, 5, count)
		assert.Equal(t, io.ErrShortWrite, errors.Cause(err))

	})

	t.Run("Context", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		var count int

		err := New(time.Millisecond).Context(ctx).Run(func() error {
			if count == 3 {
				cancel()
			} else {
				count++
			}
			return io.EOF
		})

		assert.Equal(t, 3, count)
		assert.Equal(t, io.EOF, errors.Cause(err))
	})

	t.Run("Jitter", func(t *testing.T) {
		t.Parallel()

		var durs []time.Duration
		last := time.Now()
		New(time.Millisecond * 100).Attempts(100).Jitter(0.1).Run(func() error {
			durs = append(durs, time.Since(last))
			last = time.Now()
			return io.EOF
		})

		avg := avgDurations(durs)

		if avg < time.Millisecond*90 || avg > time.Millisecond*110 {
			t.Error("bad avg dur", avg)
		}
	})

	t.Run("Log", func(t *testing.T) {
		t.Parallel()

		err := errors.New("meow")
		n := 0

		New(time.Millisecond).Log(func(e error) {
			require.Equal(t, err, e)
		}).Run(func() error {
			n++

			// indexed from one :(
			switch n {
			case 1:
				return err
			case 2:
				return nil
			}

			t.Fatal("should not reach this line.")
			panic("?")
		})
	})

	t.Run("ContinueOnNil", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		var n int
		New(time.Millisecond).ContinueOnNil().Context(ctx).Run(func() error {
			if n == 1 {
				cancel()
			}
			n++
			return nil
		})
		assert.Equal(t, 2, n)
	})
}

func avgDurations(durs []time.Duration) time.Duration {
	var sum time.Duration
	for _, dur := range durs {
		sum += dur
	}
	return sum / time.Duration(len(durs))
}
