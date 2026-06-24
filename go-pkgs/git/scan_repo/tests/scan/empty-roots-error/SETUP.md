# Scenario

**Feature**: empty roots slice returns validation error

```
len(Roots)==0 -> error: at least one root required
```

## Steps

1. Set `req.Roots` to nil/empty slice.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Roots = nil
	return nil
}
```