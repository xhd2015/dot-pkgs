# Scenario

DocumentWriteTextLimitation is non-empty and mentions write text + limits.

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "document"
	return nil
}
```
