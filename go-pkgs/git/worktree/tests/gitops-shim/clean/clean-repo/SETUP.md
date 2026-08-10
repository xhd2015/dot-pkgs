# Scenario

**Feature**: clean repo reports clean via go-pkgs IsClean and IsCleanWrk

```
# no mutations after grouping init commit
IsClean=nil, IsCleanWrk=true
```

## Preconditions

- Grouping bootstrap left the repo clean with empty porcelain.

## Steps

1. No additional mutations.

## Context

- Control leaf for clean surface after shim.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	t.Logf("clean repo at %s — no mutations", req.Dir)
	return nil
}
```
