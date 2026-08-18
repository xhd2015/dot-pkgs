# Scenario

**Feature**: `--json` prints indented `RepoResult` array

```
# multi-repo JSON output
RunCLI repo list --owner alice --json -> JSON array with matched_by
```

## Steps

1. Mock auth and `gh repo list alice` with two-repo fixture.
2. Set `req.Args` with `--json`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"repo", "list", "--owner", "alice", "--json"}
	req.GhBin = writeOwnedOnlyGh(t, "alice", fixtureFile(d, "testdata/repos.json")
	return nil
}
```