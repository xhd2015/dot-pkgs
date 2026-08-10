# Scenario

**Feature**: `LookupItems.Dirs()` — unique parent dirs of found paths

```
LookupPaths(...) -> items
items.Dirs() -> unique filepath.Dir(Path), first-seen order
```

## Steps

1. Set `Operation=dirs` (Run calls LookupPaths then Items.Dirs()).
2. Leaves arrange multi-name fixtures and assert Dirs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "dirs"
	return nil
}
```
