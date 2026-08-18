# Scenario

**Feature**: `--search-description` runs description search and formats line output

```
# description search flag
RunCLI repo list --search-description widget -> gh search repos -> description lines
```

## Steps

1. Mock auth and `gh search repos` with two-repo fixture.
2. Set `req.Args` to `["repo", "list", "--search-description", "widget"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"repo", "list", "--search-description", "widget"}
	req.GhBin = writeSearchReposGh(t, fixtureFile(d, "testdata/search-repos.json"))
	return nil
}
```