# Scenario

**Feature**: scoped tag patch segment increments

```
sub/v0.2.9 -> IncrementTag -> sub/v0.2.10
```

## Steps

1. Set `req.Tag` to `sub/v0.2.9`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Tag = "sub/v0.2.9"
	return nil
}
```