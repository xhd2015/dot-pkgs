# Scenario

**Feature**: tags that peel to the old SHA are retargeted

```
tag --points-at old -> delete local -> maybe remote --delete -> retag new -> push
```

## Preconditions

- Tag name is `v0.0.333` (motivating case).
- Remote ops use a local bare repo. Default remote is `origin` unless
  `--remote` is set.
- Remote `--delete` only if `ls-remote` still shows that tag at the old commit.

## Steps

1. Set `req.TagName` to `v0.0.333`.
2. Leaf creates the tag type + remote state and appends `-m`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.TagName = "v0.0.333"
	return nil
}
```
