# retry

An expressive, flexible retry package for Go.

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/coder/retry)

```
go get github.com/coder/retry
```

## Features
- For loop experience instead of closures
- Only 4 exported methods

## Examples

Wait for connectivity to google.com, checking at most once every
second.
```go
func pingGoogle(ctx context.Context) error {
    r := retry.New(time.Second, time.Second*10)
    for r.Wait(ctx) {
        _, err := http.Get("https://google.com")
        r.SetError(err)
    }
    return r.Error()
}
```

Wait for connectivity to google.com, checking at most 10 times.
```go
func pingGoogle(ctx context.Context) error {
    r := retry.New(time.Second, time.Second*10)
    for n := 0; r.Wait(ctx) && n < 10; n++ {
        _, err := http.Get("https://google.com")
        r.SetError(err)
    }
    return r.Error()
}
```