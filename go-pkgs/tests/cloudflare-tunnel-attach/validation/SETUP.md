# Scenario

**Feature**: Attach rejects invalid inputs before mutating managed registry

```
# validation branch
Domain empty
  -> Attach error
  -> no requirement that state.json exists
```

## Preconditions

- Lifecycle merge / connector start is out of scope for this branch.
- Leaves set Domain (or Sequence) to invalid values and ExpectError.

## Steps

1. Append `validation` to DecisionPath.
2. Descend into concrete rejection leaves.

## Context

- Highest significance under this tree: error vs success outcome class.
- Requirement scenario 4: Domain required.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "validation")
	return nil
}
```
