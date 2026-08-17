# Scenario

**Feature**: bare `go` runs `$GOROOT/bin/go` with GOROOT and PATH set

```
# fake $GOROOT/bin/go prints GOROOT, first PATH entry, WITHGO_EXTRA
args=["go"] + ExtraEnv -> Exec -> child script stdout
```

## Steps

1. Write a fake `$GOROOT/bin/go` script.
2. Set args to `["go"]` and ExtraEnv `WITHGO_EXTRA=from-test`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	writeFakeGo(t, req.Goroot)
	req.Args = []string{"go"}
	req.ExtraEnv = []string{"WITHGO_EXTRA=from-test"}
	return nil
}
```
