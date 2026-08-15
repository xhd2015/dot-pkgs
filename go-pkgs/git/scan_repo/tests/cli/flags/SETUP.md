# Scenario

**Feature**: scan flags map to `Options` with default lines output

```
caller --root [...] [flags] -> RunCLI -> Scan -> tab-separated lines
```

## Preconditions

- Fake `.git` fixtures; no real `git` required.
- `--json` not set (default lines format).

## Steps

1. Build workspace fixtures per leaf.
2. Populate `req.Args` with `--root` and optional tuning flags.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Flags branch uses default lines output — strip --json if present upstream.
	var filtered []string
	for _, arg := range req.Args {
		if arg == "--json" {
			continue
		}
		filtered = append(filtered, arg)
	}
	req.Args = filtered
	return nil
}
```