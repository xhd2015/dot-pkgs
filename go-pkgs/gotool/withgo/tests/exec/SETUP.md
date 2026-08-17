# Scenario

**Feature**: Exec runs a child with GOROOT and PATH taken from the given goroot

```
# child env; bare go resolved under goroot/bin; empty args run env
goroot + args -> Exec -> child GOROOT=$abs PATH=$abs/bin:$PATH
```

## Steps

1. Set `req.Op` to `exec`.
2. Create an isolated goroot (`t.TempDir()`).
3. Leaf `Setup` writes a fake `bin/go` and/or sets args.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "exec"
	req.Goroot = t.TempDir()
	return nil
}
```
