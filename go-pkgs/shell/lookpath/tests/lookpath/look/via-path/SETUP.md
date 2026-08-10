# Scenario

**Feature**: bare name resolved via injectable `opts.LookPath` → `Via=path`

```
LookPath(name) hit -> Path + Via=path
later stages not required
```

## Steps

1. Group marks bare-name PATH stage; leaf injects LookPathHit.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Name = "mytool"
	return nil
}
```
