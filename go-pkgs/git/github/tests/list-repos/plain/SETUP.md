# Scenario

**Feature**: plain mode lists owned repos without search

```
# no search keywords
ListRepos SearchDescription="" SearchCode="" -> ListOwned per owner -> matched_by ["owned"]
```

## Steps

1. Clear search keywords for plain-mode leaves.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SearchDescription = ""
	req.SearchCode = ""
	return nil
}
```