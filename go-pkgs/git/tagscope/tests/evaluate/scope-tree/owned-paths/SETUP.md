# Scenario

**Feature**: OwnedPathsForScope excludes nested child scope subtrees

```
scope + ScopeTree + all repo paths -> OwnedPathsForScope -> owned prefixes
```

## Steps

1. Set `req.Op` to `"owned-paths"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "owned-paths"
	return nil
}
```