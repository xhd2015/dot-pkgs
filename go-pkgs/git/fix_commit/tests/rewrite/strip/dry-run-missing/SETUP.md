# Scenario

**Feature**: missing trailer hard-errors even under `--dry-run`

```
no trailer + --strip-co-author --dry-run -> Error: commit message has no Co-authored-by line
```

## Steps

1. Amend HEAD to `fix typo` (no trailer).
2. Append `--strip-co-author` `--dry-run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	setHEADMessage(t, req.Dir, "fix typo\n")
	fillCommitMeta(t, req)
	req.Args = append(req.Args, "--strip-co-author", "--dry-run")
	return nil
}
```
