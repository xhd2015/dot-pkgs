# Scenario

**Feature**: `--json` emits JSON array of Repo structs

```
alpha + zebra repos -> JSON array with RepoType "main"
```

## Steps

1. Create workspace with `alpha/` and `zebra/` repos.
2. Set `req.Args` to `["--root", <workspace>, "--json"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	root := multiRepoWorkspace(t)
	req.Args = []string{"--root", root, "--json"}
	return nil
}
```