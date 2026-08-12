# Scenario

**Feature**: multi-GOPATH from bash — return first segment only (Clean)

```
RunLogin("bash", …) <- GOPATH=/tmp/a:/tmp/b\0
ResolveGoPathWith -> ("/tmp/a", nil)
# a:b → filepath.Clean(first segment)
```

## Steps

1. Inject bash dump with multi-path `GOPATH=/tmp/a:/tmp/b`.
2. Expect cleaned first segment only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BashStdout = nulEnvDump("GOPATH=/tmp/a:/tmp/b")
	return nil
}
```
