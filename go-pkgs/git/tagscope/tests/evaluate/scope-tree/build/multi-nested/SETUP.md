# Scenario

**Feature**: BuildScopeTree links nested scopes under parents

```
root + sub/ + sub/nested/ tags -> Children map with nested hierarchy
```

## Steps

1. Set tag names spanning three nested scopes.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"v0.0.1", "sub/v0.2.1", "sub/nested/v0.1.1"}
	return nil
}
```