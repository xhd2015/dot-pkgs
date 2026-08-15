# Scenario

**Feature**: pure parse helpers for repo naming and URL normalization

```
# no gh involved
NormalizeRepoURL(owner,name,raw) -> https://github.com/owner/name
owner + "/" + name -> FullName
```

## Preconditions

- Parse leaves do not set `req.Owners` or `req.GhBin`.
- `Run` dispatches to `NormalizeRepoURL` or FullName construction.

## Steps

1. Leaf `Setup` sets parse-specific `Request` fields.
2. `Run` returns normalized URL or FullName in `Response`.

## Context

- `normalize-url` reads multiple cases from `testdata/cases.tsv`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Owners = nil
	req.GhBin = ""
	req.IncludeForks = true
	return nil
}```