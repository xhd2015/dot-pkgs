# Scenario

**Feature**: default line output is tab-separated full_name and matched_by

```
# owned repos, no --json
RunCLI repo list --owner alice -> alice/alpha\towned lines
```

## Steps

1. Mock auth and `gh repo list alice` with two-repo fixture.
2. Set `req.Args` to `["repo", "list", "--owner", "alice"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"repo", "list", "--owner", "alice"}
	req.GhBin = writeOwnedOnlyGh(t, "alice", "testdata/repos.json")
	return nil
}
```