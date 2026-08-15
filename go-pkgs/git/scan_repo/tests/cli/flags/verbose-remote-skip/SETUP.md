# Scenario

**Feature**: `-v` prints a warning when a CloudStorage directory is skipped during walk

```
# cloud-sync subtree under walk root
caller --root + -v -> SkipDir on CloudStorage -> stderr warning
```

## Steps

1. Build workspace with local `visible-repo` and CloudStorage repo.
2. Run CLI with `--root` and `-v`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	cloudStorageProvider(t, root, "GoogleDrive-user@example.com")
	cloudRepo := filepath.Join(root, "Library", "CloudStorage", "GoogleDrive-user@example.com", "Projects", "cloud-app")
	mkdirAll(t, cloudRepo)
	fakeGitRepo(t, cloudRepo)
	visible := filepath.Join(root, "visible-repo")
	mkdirAll(t, visible)
	fakeGitRepo(t, visible)
	req.Args = []string{"--root", root, "-v"}
	return nil
}
```