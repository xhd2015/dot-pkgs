# Scenario

**Feature**: BuildScopeTree derives parent/child scope relationships

```
Collected.Scopes -> BuildScopeTree -> Children map
```

## Steps

1. Set `req.Op` to `"build-scope-tree"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "build-scope-tree"
	return nil
}
```