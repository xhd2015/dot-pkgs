# Scenario

Live e2e grouping (label e2e on leaves). Opens iTerm ForceNew windows.

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	return nil
}
```
