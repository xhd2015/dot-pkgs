# Scenario

**Feature**: `ListRepos` option forwarding (limit, multi-owner)

```
# options layer
ListReposOptions Limit / Owners -> gh argv and merged results
```

## Steps

1. Clear search keywords so options leaves exercise plain owned mode unless overridden.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SearchDescription = ""
	req.SearchCode = ""
	return nil
}
```