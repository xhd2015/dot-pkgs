# Scenario

**Feature**: code search dedupes hits and tags `matched_by: ["code"]`

```
# code search with duplicate repository hits
ListRepos SearchCode=handler -> gh search code handler -> dedupe by FullName
```

## Steps

1. Set `req.SearchCode` to `handler`.
2. Mock `gh search code` returns duplicate `tool-a` hits plus `tool-b`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SearchDescription = ""
	req.SearchCode = "handler"
	req.GhBin = writeSearchCodeGh(t, "testdata/search-code.json")
	return nil
}
```