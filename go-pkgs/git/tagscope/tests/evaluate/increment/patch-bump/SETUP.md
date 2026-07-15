# Scenario

**Feature**: patch segment rolls from 9 to 10

```
v0.0.9 -> IncrementTag -> v0.0.10
```

## Steps

1. Set `req.Tag` to `v0.0.9`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Tag = "v0.0.9"
	return nil
}
```