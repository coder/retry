package retry

import (
	"net"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type testListener struct {
	acceptFn func() (net.Conn, error)
}

func newTestListener(acceptFn func() (net.Conn, error)) net.Listener {
	return &Listener{
		LogTmpErr: func(err error) {},
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
