# Scenario

**Feature**: parent scope excludes paths owned by nested child scope

```
sub/ scope + paths under sub/nested/ -> only direct sub/ paths owned
```

## Steps

1. Set tag names with `sub/` and `sub/nested/` scopes.
2. Set `ScopePrefix` to `sub/` and `AllPaths` spanning both levels.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"sub/v0.2.1", "sub/nested/v0.1.1"}
	req.ScopePrefix = "sub/"
	req.AllPaths = []string{
		"sub/a.txt",
		"sub/pkg.go",
		"sub/nested/mod.go",
		"sub/nested/deep/x.go",
	}
	return nil
}
```