# Scenario

**Feature**: `RunCLI` routing and flag parse errors surface on stderr

```
# unknown command, subcommand, flag, or extra args
RunCLI invalid argv -> error message -> stderr, non-zero exit
```

## Preconditions

- Error leaves do not configure mock `gh`; `ListRepos` must not run.

## Steps

1. Descendant `Setup` sets `req.Args` to the invalid argv under test.

## Context

- Top-level unknown commands mention `unrecognized`.
- `repo` unknown subcommands mention `unrecognized` repo command.
- Unknown flags and trailing positionals produce parse errors.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GhBin = ""
	return nil
}
```