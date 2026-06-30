# Scenario

**Feature**: default CLI scan silently skips CloudStorage subtrees

```
# cloud-sync subtree under walk root without -v
caller --root -> SkipDir on CloudStorage -> empty stderr
```

## Steps

1. Build workspace with local `visible-repo` and CloudStorage repo.
2. Run CLI with `--root` only.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	cloudStorageProvider(t, root, "GoogleDrive-user@example.com")
	cloudRepo := filepath.Join(root, "Library", "CloudStorage", "GoogleDrive-user@example.com", "Projects", "cloud-app")
	mkdirAll(t, cloudRepo)
	fakeGitRepo(t, cloudRepo)
	visible := filepath.Join(root, "visible-repo")
	mkdirAll(t, visible)
	fakeGitRepo(t, visible)
	req.Args = []string{"--root", root}
	return nil
}
```