# Scenario

**Feature**: description and/or code search modes via `gh search`

```
# search branches
ListRepos SearchDescription and/or SearchCode -> gh search repos/code -> matched_by tags
```

## Steps

1. Default owner `alice` when leaf does not set explicit owners.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if len(req.Owners) == 0 {
		req.Owners = []string{"alice"}
	}
	return nil
}
```