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
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Owners = []string{"alice", "bob"}
	req.GhBin = writeOwnerFixtureGh(t, d, map[string]string{
		"alice": fixtureFile(d, "testdata/alice.json"),
		"bob":   fixtureFile(d, "testdata/bob.json"),
	})
	return nil
}```