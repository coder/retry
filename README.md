# retry

An expressive, flexible retry package for Go.

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/codercom/retry)

## Features

- Any combination of:
  - Exponential backoff
  - Jitter
  - Bound to context
  - Bound to timeout
  - Limit total number of attempts
  - Retry only on certain kinds of errors
    - Defaults to retrying on all errors
- Retrying net.Listener wrapper

## Examples

This code will succeed after about a second.

```go
start := time.Now()

err := retry.New(time.Millisecond).
    Backoff(time.Millisecond * 50).
    Timeout(time.Second * 2).
    Run(
        func() error {
            if time.Since(start) < time.Second {
                return errors.New("not enough time has elapsed")
            }
            return nil
        },
    )
fmt.Printf("err: %v, took %v\n", err, time.Since(start))
```

---

This code will block forever, since no `Timeout` or `Context` option
was provided.

```go
start := time.Now()

err := retry.New(time.Millisecond).
    Backoff(time.Millisecond * 50).
    Run(
        func() error {
            return errors.New("not enough time has elapsed")
        },
    )
fmt.Printf("err: %v, took %v\n", err, time.Since(start))
```

---

This code will sleep anywhere from 500ms to 1.5s between attempts.

The attempts condition will fail before the timeout. It will return the
`not enough time...` error after 2.5 to 7.5 seconds.

```go
start := time.Now()

err := retry.New(time.Second).
    Jitter(0.5).
    Timeout(time.Second * 10).
    Attempts(5).
    Run(
        func() error {
            return errors.New("not enough time has elapsed")
        },
    )
fmt.Printf("err: %v, took %v\n", err, time.Since(start))
```

## We're Hiring!

If you're a passionate Go developer, send your resume and/or GitHub link to [jobs@coder.com](mailto:jobs@coder.com).
