# Scenario

**Feature**: escape double quotes in paths

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "escape-path"
	req.EscapeInput = `/tmp/"proj"`
	return nil
}
```