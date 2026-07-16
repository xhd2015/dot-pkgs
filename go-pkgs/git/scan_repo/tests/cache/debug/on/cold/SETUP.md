# Scenario

**Feature**: first Scan with empty mirror and Debug logs cold mode

```
# cold debug (missing root cache entry)
workspace/my-repo/.git + fresh CacheRoot (no entry.json)
  -> Scan(Debug=true, NoCache=false, CacheRoot)
  -> full cold walk
  -> stderr: scan: … mode=cold … reason≈missing_root_entry
```

## Preconditions

- `CacheRoot` is a brand-new temp dir (no prior seed) → cold path.
- `Debug=true` from parent `on/`.
- `NoCache=false` so cache is enabled (cold write still allowed).

## Steps

1. Create workspace with one fake main repo `my-repo/`.
2. Set `req.Roots`; leave cache empty (do not cold-seed).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "my-repo")
	mkdirAll(t, repoDir)
	fakeGitRepo(t, repoDir)
	req.Roots = []string{root}
	req.NoCache = false
	// Debug already true from on/; CacheRoot empty of mirror entries.
	return nil
}
```
