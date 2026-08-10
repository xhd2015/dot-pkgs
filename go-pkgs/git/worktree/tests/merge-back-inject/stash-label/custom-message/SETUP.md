# Scenario

**Feature**: custom StashLabel appears in stash history after dirty-diverged migrate

```
# after successful rebased-and-merged, stash reflog still records the inject label
StashLabel=doctest-inject-stash-label -> stash push -m ... -> pop
  -> git reflog show stash contains label
```

## Preconditions

- Grouping set `req.StashLabel` to `doctest-inject-stash-label`.
- Source dirty diverged; Remove=false.

## Steps

1. Run MergeBack.
2. Assert rebased-and-merged and dirt restored.
3. Assert `git reflog show stash` (or equivalent stash history) contains the
   inject label — proves push used `StashLabel`, not a hard-coded product string.

## Context

- Stash list is empty after successful pop; **reflog** retains the message.
- If implementer ignores `StashLabel` and keeps `"wrk-merge-back"`, this leaf RED.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if req.SourcePath == "" {
		return fmt.Errorf("custom-message: ancestor fixture missing SourcePath")
	}
	if req.StashLabel != "doctest-inject-stash-label" {
		return fmt.Errorf("custom-message: want StashLabel doctest-inject-stash-label, got %q", req.StashLabel)
	}
	return nil
}
```
