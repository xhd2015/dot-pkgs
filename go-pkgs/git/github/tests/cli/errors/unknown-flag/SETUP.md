# Scenario

**Feature**: unknown `repo list` flag rejected

```
# invalid flag
RunCLI repo list --nope -> flag parse error -> stderr
```

## Steps

1. Set `req.Args` to `["repo", "list", "--nope"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"repo", "list", "--nope"}
	return nil
}
```