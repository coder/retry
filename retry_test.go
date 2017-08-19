package retry

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestAttempts(t *testing.T) {
	t.Run("Respects count and sleeps between attempts", func(t *testing.T) {
		count := 0
		start := time.Now()

		Attempts(time.Millisecond, 5, func() error {
			count++
			return errors.Errorf("asdfasdf")
		})

		assert.Equal(t, 5, count)
		assert.WithinDuration(t, start.Add(time.Millisecond*5), time.Now(), time.Millisecond*1)
	})

	t.Run("returns as soon as error is nil", func(t *testing.T) {
		start := time.Now()
		Attempts(time.Minute, 100, func() error {
			return nil
		})
		assert.WithinDuration(t, time.Now(), start, time.Millisecond)
	})
}

func TestTimeout(t *testing.T) {
	t.Run("Respects timeout and sleeps between attempts", func(t *testing.T) {
		count := 0
		start := time.Now()

		// The timing here is a little sketchy.
		Timeout(time.Millisecond, time.Millisecond*5, func() error {
			count++
			return errors.Errorf("asdfasdf")
		})

		assert.Equal(t, 5, count)
		assert.WithinDuration(t, start.Add(time.Millisecond*5), time.Now(), time.Millisecond*1)
	})

	t.Run("returns as soon as error is nil", func(t *testing.T) {
		start := time.Now()
		Timeout(time.Minute, time.Hour, func() error {
			return nil
		})
		assert.WithinDuration(t, time.Now(), start, time.Millisecond)
	})
}
