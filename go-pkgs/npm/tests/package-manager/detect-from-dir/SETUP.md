# Scenario

**Feature**: `DetectFromDir` resolves manager from project root indicators

```
# lockfiles, packageManager field, defaults
projectDir fixtures -> DetectFromDir -> Trace.Manager
```

## Steps

1. Leaf `Setup` writes temp project files and sets `req.ProjectDir`.
2. `req.Op` is `detect-dir`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "detect-dir"
	return nil
}
```