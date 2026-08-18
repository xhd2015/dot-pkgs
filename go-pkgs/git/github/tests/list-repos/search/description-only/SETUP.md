# Scenario

**Feature**: description search tags results with `matched_by: ["description"]`

```
# description search only
ListRepos SearchDescription=widget -> gh search repos widget --owner alice
```

## Steps

1. Set `req.SearchDescription` to `widget`.
2. Mock `gh search repos` returns two repos from fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SearchDescription = "widget"
	req.SearchCode = ""
	req.GhBin = writeSearchReposGh(t, fixtureFile(d, "testdata/search-repos.json")
	return nil
}
```