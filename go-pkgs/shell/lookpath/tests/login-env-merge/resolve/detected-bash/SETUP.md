# Scenario

**Feature**: DetectShell returns bash — dump bash only

```
DetectShell -> "bash"
RunLogin("bash", env -0, …) -> dump
ResolveLoginEnvs -> shell="bash", envs
```

## Steps

1. Set `DetectShellResult=bash`.
2. Leaves inject bash dump or BashFail.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.DetectShellResult = "bash"
	return nil
}
```
