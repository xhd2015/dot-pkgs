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
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SearchDescription = "widget"
	req.SearchCode = ""
	req.GhBin = writeSearchReposGh(t, "testdata/search-repos.json")
	return nil
}
```