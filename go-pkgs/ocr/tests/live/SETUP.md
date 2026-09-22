# Scenario

Live e2e grouping (label e2e on leaves). Real Apple Vision OCR.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "live_file"
	return nil
}
```
