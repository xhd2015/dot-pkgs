# Scenario

**Feature**: `repo list` formats `ListRepos` results as lines or JSON

```
# default lines or --json
RunCLI repo list -> ListRepos -> format stdout
```

## Preconditions

- Output leaves use mock `gh` with authenticated alice and canned repo JSON.

## Steps

1. Descendant `Setup` configures `req.Args`, mock `gh`, and fixture data.

## Context

- Default mode: `{full_name}\t{matched_by}` per line, sorted ascending.
- JSON mode: indented array of `RepoResult` objects.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Args == nil {
		req.Args = []string{"repo", "list"}
	}
	return nil
}
```