# Scenario

**Feature**: `repo` alone prints repo-level help (not list leaf help)

```
# repo with no subcommand
RunCLI repo -> repo-level usage (list; points to repo list --help) -> stdout, exit 0
```

## Steps

1. Set `req.Args` to `["repo"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"repo"}
	return nil
}
```
