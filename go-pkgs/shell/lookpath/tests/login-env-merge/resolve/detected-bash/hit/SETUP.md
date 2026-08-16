# Scenario

**Feature**: bash detect + multi-key NUL dump → shell=bash with FOO and GOPATH

```
DetectShell -> bash
RunLogin("bash") <- FOO=1\0GOPATH=/tmp/gp\0
ResolveLoginEnvs -> ("bash", [FOO=1, GOPATH=/tmp/gp], nil)
```

## Steps

1. Inject BashStdout with FOO=1 and GOPATH=/tmp/gp.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BashStdout = nulEnvDump("FOO=1", "GOPATH=/tmp/gp")
	return nil
}
```
