# Scenario

**Feature**: CLI cache control flags map to `Options` and affect durable cache I/O

```
# --no-cache       -> Options.NoCache=true (no index/walk/mirror write)
# --cache-dir PATH -> Options.CacheRoot=PATH (index + walk under PATH; no mirror)

workspace + flags
  -> RunCLI -> Scan -> stdout lines + optional home/ under C
```

## Preconditions

- Nested CLI tree; uses its own Run.
- Leaves set `req.Args` with `--root` and cache flags.
- `fakeGitRepo` / `rootsFromArgs` come from parent `cli/SETUP.md`.

## Steps

1. Normalize empty Args; provide `cacheDirFromArgs` for asserts under this branch.

```go
import (
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.Args == nil {
		req.Args = []string{}
	}
	return nil
}

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
```
