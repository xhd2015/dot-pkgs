# Scenario

**Feature**: TellApplicationHeader path-bound vs bare fallback

```
appPath -> TellApplicationHeader -> tell line
```

## Steps

1. Phase `tell-header`.
2. Leaves set AppPath (non-empty or empty).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "tell-header"
	return nil
}
```
