# Scenario

**Feature**: a runner failure surfaces instead of a Result

```
Resolve ok -> Run -> error
```

## Steps

1. Set `Kind=dir` and `RunnerErr`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "dir"
	req.RunnerErr = "vscode: exit status 1"
	return nil
}
```
