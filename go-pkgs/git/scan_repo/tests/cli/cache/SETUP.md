# Scenario

**Feature**: CLI cache control flags map to `Options` and affect mirror I/O

```
# --cache-dir PATH -> Options.CacheRoot
# --no-cache       -> Options.NoCache=true (no mirror read/write)
# --refresh        -> Options.Refresh=true (force cold full walk + rewrite)
caller --root R [--cache-dir C] [--no-cache] [--refresh]
  -> RunCLI -> Scan -> stdout lines + optional mirror under C
```

## Preconditions

- Fake `.git` fixtures; no enrichment flags.
- Leaves always pass an explicit temp `--cache-dir` (never the product default under `$HOME`).
- Default lines output (no `--json`).

## Steps

1. Provide helpers to parse `--cache-dir` from argv for assertions.
2. Leaves build workspace, set `req.Args` with `--root` and cache flags.

```go
import (
	"strings"
	"testing"
)

func cacheDirFromArgs(args []string) string {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--cache-dir" && i+1 < len(args):
			return args[i+1]
		case strings.HasPrefix(args[i], "--cache-dir="):
			return strings.TrimPrefix(args[i], "--cache-dir=")
		}
	}
	return ""
}

func Setup(t *testing.T, req *Request) error {
	// Cache branch: leave Args empty for leaves to populate.
	if req.Args == nil {
		req.Args = []string{}
	}
	return nil
}
```
