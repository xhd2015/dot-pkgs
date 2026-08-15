# Scenario

**Feature**: login shell must not be replaced with one-shot exec

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-no-exec")
	return nil
}
```