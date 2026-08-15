# Scenario

**Feature**: empty `RunCLI` args print top-level usage (same class as `--help`)

```
# empty argv after kool github
RunCLI [] -> top-level usage mentioning repo -> stdout, exit 0
```

## Steps

1. Set `req.Args` to `[]string{}` (empty).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{}
	return nil
}
```
