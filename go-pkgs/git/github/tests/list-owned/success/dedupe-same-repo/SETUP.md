# Scenario

**Feature**: duplicate FullName from two owner queries keeps first occurrence

```
# dedupe by FullName (first owner in opts.Owners wins)
ListOwned owners=[alice,bob] -> same alice/shared twice -> one Repo
```

## Steps

1. Mock returns `alice/shared` for both `alice` and `bob` owner queries.
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