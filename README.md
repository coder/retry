# retry

An expressive, flexible retry package for Go.

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/codercom/retry)

## Features

- Exponential backoff
- Jitter
- Bound to context
- Bound to timeout
- Bound to attempt count
- Bound to certain kinds of errors
- Any combination of the above
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
