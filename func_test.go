package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFunc(t *testing.T) {
	t.Parallel()

	passAfter := time.Now().Add(time.Second)

	dog, err := Func(func() (string, error) {
		if time.Now().Before(passAfter) {
			return "", errors.New("not yet")
		}
		return "dog", nil
	},
		New(time.Millisecond, time.Millisecond))(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dog != "dog" {
		t.Fatal("expected dog")
	}
}
