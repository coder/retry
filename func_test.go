package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFunc(t *testing.T) {
	t.Parallel()

	passAfter := time.Now().Add(time.Second)

	dog, err := Func[string](func() (string, error) {
		if time.Now().Before(passAfter) {
			return "", errors.New("not yet")
		}
		return "dog", nil
	}).Do(context.Background(), New(time.Millisecond, time.Second))
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "dog", dog)
}
