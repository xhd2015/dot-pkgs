# Scenario

**Feature**: `repo --help` prints the same repo-level help as bare `repo`

```
# repo help flags
RunCLI repo --help -> repo-level usage -> stdout, exit 0
```

## Steps

1. Set `req.Args` to `["repo", "--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"repo", "--help"}
	return nil
}
```
