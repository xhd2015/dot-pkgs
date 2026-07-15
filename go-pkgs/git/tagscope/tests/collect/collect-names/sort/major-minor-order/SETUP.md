# Scenario

**Feature**: major and minor version boundaries sort correctly

```
[v1.0.0, v0.9.9, v2.0.0] -> Tags v2.0.0, v1.0.0, v0.9.9
```

## Steps

1. Set `req.Names` spanning major version boundaries.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"v1.0.0", "v0.9.9", "v2.0.0"}
	return nil
}
```