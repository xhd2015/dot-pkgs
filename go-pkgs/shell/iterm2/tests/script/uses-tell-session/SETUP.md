# Scenario

**Feature**: session path read via tell aSession

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-tell-session")
	return nil
}
```