# Scenario

**Feature**: format helpers render entry blocks and summary without filesystem scan

```
constructed EntryResult / ScanSummary -> FormatEntryBlock / FormatSummaryLines -> text
```

## Preconditions

- Format leaves set `req.Mode` to `format-entry` or `format-summary`.
- No temp HOME or `analyse.Scan` call required.

## Steps

1. Leaves populate `req.Entry` or `req.Summary` with representative values.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Format leaves override Mode; default to entry-block for grouping descendants.
	if req.Mode == "" || req.Mode == "scan" {
		req.Mode = "format-entry"
	}
	return nil
}
```