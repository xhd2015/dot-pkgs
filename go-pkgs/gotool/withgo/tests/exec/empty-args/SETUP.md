# Scenario

**Feature**: empty args run the `env` command under the pinned GOROOT env

```
# no args -> env (kool); child still gets GOROOT and PATH prefix
args=[] -> Exec -> env; GOROOT=$abs; PATH starts with $abs/bin
```

## Steps

1. Leave goroot empty (no `bin/go`). Set `req.Args` to nil/empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Args = nil
	return nil
}
```
