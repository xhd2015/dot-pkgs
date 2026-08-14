# Scenario

**Feature**: leftover message empty after strip is fatal `Error:`

```
message is only "Co-authored-by: Bot <bot@x>" -> --strip-co-author -> Error:
```

## Steps

1. Amend HEAD so the entire message is one trailer line.
2. Append `--strip-co-author`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	setHEADMessage(t, req.Dir, "Co-authored-by: Bot <bot@x>\n")
	fillCommitMeta(t, req)
	req.Args = append(req.Args, "--strip-co-author")
	return nil
}
```
