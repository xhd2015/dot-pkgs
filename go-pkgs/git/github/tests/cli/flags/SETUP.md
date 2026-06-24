# Scenario

**Feature**: `repo list` flags map to `ListReposOptions` and affect gh invocations

```
# --owner or --search-description
RunCLI repo list --flag -> ListRepos with options -> mock gh argv/output
```

## Preconditions

- Flag leaves use mock `gh` scripts tailored to the flag under test.

## Steps

1. Descendant `Setup` sets `req.Args` with the flag and configures mock `gh`.

## Context

- `--owner` scopes `gh repo list <owner>`.
- `--search-description` triggers `gh search repos` instead of repo list.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Args == nil {
		req.Args = []string{"repo", "list"}
	}
	return nil
}
```