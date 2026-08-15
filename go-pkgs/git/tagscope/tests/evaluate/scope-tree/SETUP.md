# Scenario

**Feature**: scope tree and owned-path partitioning from collected scopes

```
CollectedTags -> BuildScopeTree -> parent/child map
scope + all paths -> OwnedPathsForScope -> owned path prefixes
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Op != "" && req.Op != "build-scope-tree" && req.Op != "owned-paths" {
		t.Fatalf("Op = %q, want build-scope-tree or owned-paths", req.Op)
	}
	return nil
}
```