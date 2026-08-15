# Scenario

**Feature**: second Scan processes gen_end 1 and appends gen_end 2

```
# workspace after cold has gen_end 1
workspace/
  notes/
  projects/alpha/  (fake .git)
  -> cold Scan seals gen_end 1
  -> second Scan(delta>=60s, WarmRefreshBudget=-1)
  -> consumer reads from cursor through gen_end 1
  -> appends {"op":"gen_end","gen":2}
  -> last gen_end in walk.jsonl has gen=2
```

## Steps

1. Plant workspace with intermediate dirs + main `projects/alpha`.
2. Keep default consume clocks (delta ≥ 60s).
3. No post-cold mutations — pure seal advancement.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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
