# Scenario

**Feature**: empty path is returned unchanged

```
path="" -> ""
```

## Steps

1. Set `req.Path` to `""` with empty env.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Env = []string{}
	req.Path = ""
	return nil
}
```
