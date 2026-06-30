# Scenario

**Feature**: `Scan` skips `Library/CloudStorage` subtrees and still discovers local repos

```
# cloud-sync path under walk root
caller roots -> Scan -> SkipDir on CloudStorage -> visible-repo row only
```

## Steps

1. Create a CloudStorage provider tree with a fake git repo inside it.
2. Create a sibling local `visible-repo` with fake `.git`.
3. Scan from the workspace root without verbose logging.

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
	req.Roots = []string{root}
	req.Verbose = false
	return nil
}
```