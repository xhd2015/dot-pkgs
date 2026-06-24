# Scenario

**Feature**: two owners return disjoint repos merged and sorted

```
# multi-owner merge
ListOwned owners=[alice,bob] -> gh x2 -> merge -> sort FullName ascending
```

## Steps

1. Mock `gh` returns `testdata/alice.json` for `alice` and `testdata/bob.json` for `bob`.
2. Set `req.Owners` to `["alice", "bob"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Owners = []string{"alice", "bob"}
	req.GhBin = writeOwnerFixtureGh(t, map[string]string{
		"alice": "testdata/alice.json",
		"bob":   "testdata/bob.json",
	})
	return nil
}```