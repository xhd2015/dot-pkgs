# Scenario

**Feature**: empty version output is unparseable

```
"" -> ParseVersion -> error
```

## Steps

1. Set `VersionOutput` to empty string.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.VersionOutput = ""
	return nil
}
```
