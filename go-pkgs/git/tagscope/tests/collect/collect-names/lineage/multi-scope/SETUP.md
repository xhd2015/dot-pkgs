# Scenario

**Feature**: multiple scopes each get independent lineage entries

```
[v0.0.1, sub/v0.0.2] -> CollectFromNames -> root and sub/ in ByScope
```

## Steps

1. Set `req.Names` spanning root and `sub/` scopes.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"v0.0.1", "sub/v0.0.2"}
	return nil
}
```