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

func Func[T any](fn func() (T, error), r *Retrier) func(context.Context) (T, error) {
	return func(ctx context.Context) (T, error) {
		var (
			v   T
			err error
		)
		for r.Wait(ctx) {
			v, err = fn()
			if err == nil {
				return v, nil
			}
			if _, ok := err.(abortError); ok {
				return v, err
			}
		}
		return v, ctx.Err()
	}
}
