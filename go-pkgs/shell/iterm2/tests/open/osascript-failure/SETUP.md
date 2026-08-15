# Scenario

**Feature**: osascript failure wrapped and returned

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = t.TempDir()
	req.OsascriptFail = true
	return nil
}
```