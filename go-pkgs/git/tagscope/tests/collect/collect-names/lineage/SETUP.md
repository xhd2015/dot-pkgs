# Scenario

**Feature**: per-scope lineage derives Newest, LatestRelease, HasPrereleaseHead

```
tags per scope -> CollectFromNames -> ScopeLineage
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Op != "collect-names" {
		t.Fatalf("Op = %q, want collect-names", req.Op)
	}
	return nil
}
```