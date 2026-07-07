# Scenario

**Feature**: `Build` transforms scan results into nested snapshot nodes

```
scan_repo.Result -> Build(rel) -> Snapshot
```

## Steps

1. Group leaves by input source (live scan vs manual result).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.BaseDir == "" {
		req.BaseDir = t.TempDir()
	}
	return nil
}
```
