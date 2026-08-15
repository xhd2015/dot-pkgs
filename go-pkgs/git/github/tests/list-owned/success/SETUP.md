# Scenario

**Feature**: `ListOwned` happy paths — mapping, merge, dedupe, empty results

```
# successful gh JSON decode and merge
ListOwned -> gh repo list <owner> -> JSON array -> []Repo sorted by FullName
```

## Preconditions

- Mock `gh` returns valid JSON for each queried owner.
- `opts.Owners` contains at least one non-empty username.

## Steps

1. Configure mock `gh` with canned JSON from leaf `testdata/`.
2. Set `req.Owners` and default options unless overridden by leaf.
3. `Run` calls `ListOwned` and returns merged `[]Repo`.

## Context

- Default `IncludeArchived` is false and `IncludeForks` is true.
- URL fields from `gh` may be SSH or https; implementation normalizes to https.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Limit == 0 {
		req.Limit = 100
	}
	req.IncludeForks = true
	return nil
}```