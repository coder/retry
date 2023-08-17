# retry

An exponentially backing off retry package for Go.

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/coder/retry)

```
go get github.com/coder/retry@latest
```

`retry` promotes control flow using `for`/`goto` instead of callbacks.

## Examples

Wait for connectivity to google.com, checking at most once every
second:

```go
func pingGoogle(ctx context.Context) error {
	var err error

	r := retry.New(time.Second, time.Second*10);

	// Jitter is useful when the majority of clients to a service use
	// the same backoff policy.
	//
	// It is provided as a standard deviation.
	r.Jitter = 0.1

  retry:
	_, err = http.Get("https://google.com")
	if err != nil {
		if r.Wait(ctx) {
			goto retry
		}
		return err
	}

	return nil
}
```

Wait for connectivity to google.com, checking at most 10 times:
```go
func pingGoogle(ctx context.Context) error {
	var err error
	
	for r := retry.New(time.Second, time.Second*10); n := 0; r.Wait(ctx) && n < 10; n++ {
		_, err = http.Get("https://google.com")
		if err != nil {
			continue
		}
		break
	}
	return err
}
```