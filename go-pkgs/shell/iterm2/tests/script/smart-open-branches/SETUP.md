# Scenario

**Feature**: smart-open branches in generated script

```
BuildScript -> scan paths -> create tab OR create window
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-doctest-proj")
	return nil
}
```