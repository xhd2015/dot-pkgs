# Scenario

**Feature**: strip all matching Co-authored-by lines (mixed case + CR)

```
two trailers (one Co-Authored-By, one with \r) -> --strip-co-author -> "fix typo"
```

## Steps

1. Default strip fixture already has the two trailers.
2. Append `--strip-co-author`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "--strip-co-author")
	return nil
}
```
