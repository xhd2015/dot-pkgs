# Scenario

**Feature**: two-component versions are rejected

```
v0.0 -> ParseTagName -> ok=false
```

## Steps

1. Set `req.Name` to `v0.0`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Name = "v0.0"
	return nil
}
```