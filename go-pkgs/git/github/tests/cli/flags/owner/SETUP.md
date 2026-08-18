# Scenario

**Feature**: `--owner` passes explicit owner to `gh repo list`

```
# explicit owner flag
RunCLI repo list --owner alice -> gh repo list alice
```

## Steps

1. Mock auth and single-repo fixture for alice.
2. Set `req.Args` to `["repo", "list", "--owner", "alice"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"repo", "list", "--owner", "alice"}
	req.GhBin = writeOwnedOnlyGh(t, "alice", fixtureFile(d, "testdata/repos.json"))
	return nil
}
```