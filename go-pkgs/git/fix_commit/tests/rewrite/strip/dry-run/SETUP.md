# Scenario

**Feature**: strip dry-run shows the leftover subject and mutates nothing

```
--strip-co-author --dry-run -> [dry-run] message: fix typo + stripped line -> HEAD unchanged
```

## Steps

1. Keep the default trailer fixture.
2. Append `--strip-co-author` `--dry-run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "--strip-co-author", "--dry-run")
	return nil
}
```
