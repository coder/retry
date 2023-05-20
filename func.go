package retry

import "context"

type abortError struct {
	error
}

// Abort returns an error that will cause the retry loop to immediately abort.
// The underlying error will be returned from the Do method.
func Abort(err error) error {
	return abortError{err}
}

// Func is a retriable function that returns a value and an error.
type Func[T any] func() (T, error)

func (f Func[T]) Do(ctx context.Context, r *Retrier) (T, error) {
	var (
		v   T
		err error
	)
	for r.Wait(ctx) {
		v, err = f()
		if err == nil {
			return v, nil
		}
		if _, ok := err.(abortError); ok {
			return v, err
		}
	}
	return v, ctx.Err()
}
