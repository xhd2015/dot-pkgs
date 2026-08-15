# Scenario

**Feature**: `repo list --help` prints list subcommand usage and flags

```
# list help
RunCLI repo list --help -> list usage with flags -> stdout
```

## Steps

1. Set `req.Args` to `["repo", "list", "--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"repo", "list", "--help"}
	return nil
}
```
