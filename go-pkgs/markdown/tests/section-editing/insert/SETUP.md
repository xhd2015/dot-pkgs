# Scenario

**Feature**: insert a new section before the first recognized section

```
Caller -> Document insert -> new Section
Document -> Caller preamble and existing sections preserved
```

## Steps

1. Select the insert operation.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "insert"
	return nil
}
```
