# Scenario

**Feature**: multiple explicit owners merged and sorted with `matched_by: ["owned"]`

```
# multi-owner plain mode
ListRepos owners=[alice,bob] -> gh repo list x2 -> merge sort -> owned tags
```

## Steps

1. Set `req.Owners` to `["alice", "bob"]`.
2. Mock per-owner fixtures from `testdata/`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Owners = []string{"alice", "bob"}
	req.GhBin = writeMultiOwnerGh(t, map[string]string{
		"alice": fixtureFile(d, "testdata/alice.json"),
		"bob":   fixtureFile(d, "testdata/bob.json"),
	})
	return nil
}
```