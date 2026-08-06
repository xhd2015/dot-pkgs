# Scenario

Exactly WriteTextSafeMaxBytes (900) ASCII → OK.

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "check"
	req.CheckExactLen = applescript.WriteTextSafeMaxBytes
	return nil
}
```
