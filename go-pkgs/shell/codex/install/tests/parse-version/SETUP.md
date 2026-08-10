# Scenario

**Feature**: `ParseVersion` — extract first semver X.Y.Z from version text

```
version command / package output -> ParseVersion -> "X.Y.Z" | error
```

## Steps

1. Set `Operation=parse-version`.
2. Leaves set `VersionOutput` and assert parsed version or error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "parse-version"
	return nil
}
```
