# Scenario

**Feature**: path equal to an env alias value displays as `$NAME` alone

```
X=/tmp/proj
path=/tmp/proj
-> $X
```

## Steps

1. Create a temp dir; set env `X=<dir>` and path to that dir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	proj := t.TempDir()
	req.Env = []string{envPair("X", proj)}
	req.Path = proj
	return nil
}
```
