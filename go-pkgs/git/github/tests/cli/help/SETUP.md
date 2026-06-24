# Scenario

**Feature**: `RunCLI` help paths print usage to stdout without calling `ListRepos`

```
# top-level or repo list help
RunCLI --help OR repo list --help -> usage text -> stdout, exit 0
```

## Preconditions

- Help leaves do not set `req.GhBin`; no `gh` invocation expected.

## Steps

1. Descendant `Setup` sets `req.Args` to the help argv for that level.

## Context

- Top-level help lists `repo` among available commands.
- `repo list --help` documents list flags (`--owner`, `--json`, search flags).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GhBin = ""
	return nil
}
```