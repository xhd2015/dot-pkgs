# Scenario

**Feature**: OpenConfig invokes injectable osascript with script

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = t.TempDir()
	return nil
}
```