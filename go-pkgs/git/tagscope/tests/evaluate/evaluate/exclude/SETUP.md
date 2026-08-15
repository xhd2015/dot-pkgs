# Scenario

**Feature**: nested scopes own disjoint path prefixes

```
change under child scope -> parent owned trees unchanged -> parent skips, child bumps
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Op != "evaluate" {
		t.Fatalf("Op = %q, want evaluate", req.Op)
	}
	return nil
}
```