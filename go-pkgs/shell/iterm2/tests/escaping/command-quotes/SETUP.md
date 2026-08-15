# Scenario

**Feature**: escape double quotes in follow-up commands

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "escape-command"
	req.EscapeInput = `echo "hi"`
	return nil
}
```