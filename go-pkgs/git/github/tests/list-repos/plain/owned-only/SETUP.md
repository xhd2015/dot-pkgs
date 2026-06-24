# Scenario

**Feature**: owned-only mode tags every repo with `matched_by: ["owned"]`

```
# explicit owner, no search
ListRepos owners=[alice] -> gh repo list alice -> 2 repos -> matched_by owned
```

## Steps

1. Mock auth and `gh repo list alice` with two-repo fixture.
2. Set `req.Owners` to `["alice"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Owners = []string{"alice"}
	req.GhBin = writeOwnedOnlyGh(t, "alice", "testdata/repos.json")
	return nil
}
```