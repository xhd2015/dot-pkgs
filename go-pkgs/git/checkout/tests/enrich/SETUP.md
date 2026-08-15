# Scenario

**Feature**: `Enrich` inspects a single checkout directory

```
repo path -> Enrich(ctx, path, opts) -> Meta (never returns error)
```

## Preconditions

- `git` on PATH (skip otherwise).

## Steps

1. Check git availability for every leaf in this branch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	gitAvailable(t)
	return nil
}
```
