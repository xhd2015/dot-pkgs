# Scenario

**Feature**: `--strip-co-author` and `-m` are mutually exclusive

```
RunCLI HEAD -m x --strip-co-author -> Error: --strip-co-author and -m cannot be used together
```

## Steps

1. Set `req.Args` to `["HEAD", "-m", "x", "--strip-co-author"]`.
2. Mutual exclusion is a parse error; no repo is created.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = []string{"HEAD", "-m", "x", "--strip-co-author"}
	return nil
}
```
