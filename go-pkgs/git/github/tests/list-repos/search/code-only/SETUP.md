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
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SearchDescription = ""
	req.SearchCode = "handler"
	req.GhBin = writeSearchCodeGh(t, "testdata/search-code.json")
	return nil
}
```