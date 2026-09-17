# Scenario

**Feature**: `WellKnownCandidates` is pure — it only builds paths

```
WellKnownCandidates(home) -> []string
```

## Steps

1. Set `Operation=candidates`.
2. Leaves set `Home`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "candidates"
	return nil
}
```
