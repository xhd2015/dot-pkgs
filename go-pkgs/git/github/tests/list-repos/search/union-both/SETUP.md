# Scenario

**Feature**: description and code search union merges `matched_by`

```
# both search modes OR-merged by FullName
ListRepos SearchDescription=term SearchCode=term -> union -> matched_by merged
```

## Steps

1. Set both `req.SearchDescription` and `req.SearchCode` to `term`.
2. Mock returns overlapping and disjoint repos from description vs code fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SearchDescription = "term"
	req.SearchCode = "term"
	req.GhBin = writeUnionSearchGh(t, "testdata/search-repos.json", "testdata/search-code.json")
	return nil
}
```