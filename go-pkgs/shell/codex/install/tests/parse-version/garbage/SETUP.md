# Scenario

**Feature**: garbage with no semver is unparseable

```
"not-a-version" -> ParseVersion -> error
```

## Steps

1. Set `VersionOutput` to text without an `X.Y.Z` semver.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.VersionOutput = "not-a-version"
	return nil
}
```
