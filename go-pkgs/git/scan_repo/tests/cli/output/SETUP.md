# Scenario

**Feature**: output format selection (lines vs JSON)

```
Scan results -> lines (default) or JSON (--json) on stdout
```

## Preconditions

- Uses multi-repo or empty fixtures depending on leaf.
- No enrichment flags unless noted.

## Steps

1. Build workspace fixture shared by format leaves.
2. Set `--json` flag only on JSON leaves.

```go
import (
	"path/filepath"
	"testing"
)

func multiRepoWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"zebra", "alpha"} {
		dir := filepath.Join(root, name)
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}
	return root
}

func Setup(t *testing.T, req *Request) error {
	// Output branch leaves set --root via multiRepoWorkspace or empty temp dir.
	if len(rootsFromArgs(req.Args)) == 0 {
		req.Args = append(req.Args, "--root", multiRepoWorkspace(t))
	}
	return nil
}
```