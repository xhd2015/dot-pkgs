# Scenario

**Feature**: `co-authored by` without a colon is not a trailer

```
message contains "co-authored by Alice" -> --strip-co-author -> same missing-trailer error
```

## Steps

1. Amend HEAD to a body that mentions co-authoring without a colon.
2. Append `--strip-co-author`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	setHEADMessage(t, req.Dir, "fix typo\n\nco-authored by Alice\n")
	fillCommitMeta(t, req)
	req.Args = append(req.Args, "--strip-co-author")
	return nil
}
```
