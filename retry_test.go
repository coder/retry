package retry

import (
	"testing"
	"time"

	"net"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type testListener struct {
	acceptFn func() (net.Conn, error)
}

func newTestListener(acceptFn func() (net.Conn, error)) net.Listener {
	return &Listener{
		LogTmpErr: func(err error, retry time.Duration) {},
		Listener: &testListener{
			acceptFn: acceptFn,
		},
	}
}

func (l *testListener) Accept() (net.Conn, error) {
	return l.acceptFn()
}

func (l *testListener) Close() error {
	panic("do not call")
}

func (l *testListener) Addr() net.Addr {
	panic("do not call")
}

type testNetError struct {
	temporary bool
}

func (e *testNetError) Error() string {
	return "test net error"
}

func (e *testNetError) Temporary() bool {
	return e.temporary
}

func (e *testNetError) Timeout() bool {
	panic("do not call")
}

func TestListener(t *testing.T) {
	t.Parallel()
	t.Run("general error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("general error")
		acceptFn := func() (net.Conn, error) {
			return nil, expectedErr
		}

		_, err := newTestListener(acceptFn).Accept()
		require.Equal(t, expectedErr, err)
	})
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		acceptFn := func() (net.Conn, error) {
			return nil, nil
		}

		_, err := newTestListener(acceptFn).Accept()
		require.Nil(t, err)
	})
	t.Run("non temp net error", func(t *testing.T) {
		t.Parallel()

		expectedErr := &testNetError{false}
		acceptFn := func() (net.Conn, error) {
			return nil, expectedErr
		}

		_, err := newTestListener(acceptFn).Accept()
		require.Equal(t, expectedErr, err)
	})
	t.Run("3x temp net error", func(t *testing.T) {
		t.Parallel()

		callCount := 0
		acceptFn := func() (net.Conn, error) {
			callCount++
			switch callCount {
			case 1:
				return nil, &testNetError{true}
			case 2:
				return nil, &testNetError{true}
			case 3:
				return nil, nil
			default:
				t.Fatal("test listener called too many times; callCount: %v", callCount)
				panic("unreachable")
			}
		}

		_, err := newTestListener(acceptFn).Accept()
		require.Nil(t, err)
		require.Equal(t, callCount, 3)
	})
}
