# Scenario

**Feature**: root scope excludes paths under child scope prefixes

```
root scope + paths under sub/ -> only top-level paths owned by root
```

## Steps

1. Set tag names with root and `sub/` scopes.
2. Set `ScopePrefix` to root and `AllPaths` spanning root and child paths.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"v0.0.1", "sub/v0.2.1"}
	req.ScopePrefix = ""
	req.AllPaths = []string{
		"README",
		"go.mod",
		"sub/a.txt",
		"sub/nested/b.txt",
	}
	return nil
}
```