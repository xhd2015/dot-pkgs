# Scenario

**Feature**: default tab-separated lines output, path-sorted

```
alpha + zebra repos -> alpha\tmain\nzebra\tmain
```

## Steps

1. Create workspace with `alpha/` and `zebra/` repos.
2. Set `req.Args` to `["--root", <workspace>]` (no `--json`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	root := multiRepoWorkspace(t)
	req.Args = []string{"--root", root}
	return nil
}
```