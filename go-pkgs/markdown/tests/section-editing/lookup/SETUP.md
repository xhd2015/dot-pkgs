# Scenario

**Feature**: inspect a Markdown section without changing the document

```
Caller -> Document lookup -> Section body or selector outcome
Document -> Caller exact original bytes
```

## Steps

1. Select the lookup operation.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "lookup"
	return nil
}
```
