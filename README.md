# retry

An exponentially backing off retry package for Go.

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/coder/retry)

```
go get github.com/coder/retry
```

## Features
- A `for` loop experience instead of closures
- Only 2 exported methods
- No external dependencies

## Examples

Wait for connectivity to google.com, checking at most once every
second.
```go
func pingGoogle(ctx context.Context) error {
	var err error
	
	for r := retry.New(time.Second, time.Second*10); r.Wait(ctx); {
		_, err = http.Get("https://google.com")
		if err != nil {
			continue
		}
		break
	}
	return err
}
```

Wait for connectivity to google.com, checking at most 10 times.
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