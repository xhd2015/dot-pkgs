# Scenario

**Feature**: `--strip-co-author` with no matching line is fatal

```
message "fix typo" (no trailer) -> --strip-co-author -> Error: commit message has no Co-authored-by line
```

## Steps

1. Amend HEAD back to `fix typo` (no trailer).
2. Append `--strip-co-author`.

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
	req.Args = append(req.Args, "--strip-co-author")
	return nil
}
```
