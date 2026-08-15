# Scenario

**Feature**: non-matching tag names are rejected by `ParseTagName`

```
invalid tag name -> ParseTagName -> ok=false
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Op != "parse" {
		t.Fatalf("Op = %q, want parse", req.Op)
	}
	return nil
}
```