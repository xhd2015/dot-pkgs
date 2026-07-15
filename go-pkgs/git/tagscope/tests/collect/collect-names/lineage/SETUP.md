# Scenario

**Feature**: per-scope lineage derives Newest, LatestRelease, HasPrereleaseHead

```
tags per scope -> CollectFromNames -> ScopeLineage
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Op != "collect-names" {
		t.Fatalf("Op = %q, want collect-names", req.Op)
	}
	return nil
}
```