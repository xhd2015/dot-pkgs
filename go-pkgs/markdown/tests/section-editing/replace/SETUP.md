# Scenario

**Feature**: replace only a selected section body

```
Caller -> Document replace -> Section body changes
Document -> Caller unrelated bytes preserved
```

## Steps

1. Select the replace operation.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "replace"
	return nil
}
```
