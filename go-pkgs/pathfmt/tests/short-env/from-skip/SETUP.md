# Scenario

**Feature**: ineligible env entries are ignored even when path sits under their value

```
# skip branch
PATH/PWD/secrets/HOME/home-valued vars present -> not used as $NAME
-> fallback TildeHome or absolute
```

## Preconditions

- Leaves place path under a would-be alias value that must be skipped.
- Asserts prove the display is **not** `$SKIPPED_NAME...`.

## Steps

1. Set `req.Op = "from"`.
2. Leaves inject skipped keys and paths under those values.

## Context

- Skip rules are eligibility filters before longest-prefix selection.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "from"
	return nil
}
```
