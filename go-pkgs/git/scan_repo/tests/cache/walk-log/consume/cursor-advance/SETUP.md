# Scenario

**Feature**: walk cursor advances past cold EOF after consume seals gen_end 2

```
# cursor lifecycle
cold -> walk.cursor.json offset = len(walk.jsonl) = ColdCursorOffset
  -> second Scan consumes + appends gen_end 2 (+ optional re-list noise)
  -> walk.cursor.json offset = new len(walk.jsonl) > ColdCursorOffset
```

## Steps

1. Same fixture as seal-gen-end-2 (notes + projects/alpha).
2. Default delta ≥ 60s clocks.
3. Assert focuses on cursor math, not repo discovery details.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "notes"))
	alpha := filepath.Join(root, "projects", "alpha")
	mkdirAll(t, alpha)
	fakeGitRepo(t, alpha)
	req.Roots = []string{root}
	req.Consume = true
	return nil
}
```
