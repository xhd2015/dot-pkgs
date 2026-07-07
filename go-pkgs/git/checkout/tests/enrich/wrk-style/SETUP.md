# Scenario

**Feature**: wrk enrichment uses FormatWrk on a clean repo

```
clean repo + StatusStyle FormatWrk -> Enrich -> Status: clean
```

## Steps

1. Create committed repo on `main`.
2. Set `StatusStyle` to `FormatWrk` and `PorcelainUntracked` to false.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "main")
	gitInitRepo(t, repoDir)
	gitInitialCommit(t, repoDir, "main", "wrk fixture")
	req.RepoPath = repoDir
	req.StatusStyle = status.StyleWrk
	untracked := false
	req.PorcelainUntracked = &untracked
	return nil
}
```