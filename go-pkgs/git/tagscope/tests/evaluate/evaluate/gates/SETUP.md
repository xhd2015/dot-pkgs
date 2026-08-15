# Scenario

**Feature**: Evaluate skips when a gate rule fires before diff

```
scope lineage + commits + trees -> gate check -> SkipReason set, NextTag empty
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Op != "evaluate" {
		t.Fatalf("Op = %q, want evaluate", req.Op)
	}
	return nil
}
```