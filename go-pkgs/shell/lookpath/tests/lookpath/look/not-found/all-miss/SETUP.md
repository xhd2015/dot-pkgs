# Scenario

**Feature**: complete miss for bare name `mytool`

```
all stages miss -> Look returns error mentioning "mytool"
```

## Steps

1. Rely on parent not-found fixtures (no files, login fails).
2. Keep `Name=mytool`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Parent already configured a full miss; keep Name explicit.
	req.Name = "mytool"
	return nil
}
```
