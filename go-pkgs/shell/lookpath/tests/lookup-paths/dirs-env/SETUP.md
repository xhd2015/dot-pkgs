# Scenario

**Feature**: `LookupItems.DirsEnv()` — join Dirs with PathListSeparator

```
LookupPaths(...) -> items
items.DirsEnv() -> strings.Join(Dirs(), string(os.PathListSeparator))
                 or "" when no dirs
```

## Steps

1. Set `Operation=dirs-env` (Run calls LookupPaths then Items.DirsEnv()).
2. Leaves arrange multi-dir fixtures and assert DirsEnv string.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "dirs-env"
	return nil
}
```
