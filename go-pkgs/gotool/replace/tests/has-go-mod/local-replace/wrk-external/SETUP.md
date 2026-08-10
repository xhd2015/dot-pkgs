# Scenario

**Feature**: local replace target inside the scanning worktree's directory tree

```
# hook/guard scope = current worktree toplevel (main or linked, external/ or not)
# a replace is allowed when its resolved target lies under that worktree root
# nested linked worktrees under worktree/external/ have their own git toplevel
# but are still physically inside the parent worktree — must not be extra-repo
worktree -> external/dep linked wt + abs replace under worktree -> within-worktree
```

## Preconditions

- Consumer and dep are separate git repositories.
- The dep worktree is linked under `{worktree}/external/mydep` via `git worktree add`.
- Consumer `go.mod` has an absolute-path replace to the external path (as `wrk --dep` writes).

## Steps

1. Initialize consumer and dep git repos.
2. Add dep worktree under `consumer/external/mydep`.
3. Write consumer `go.mod` with absolute replace to the external path.
4. Call `replace.CheckLocalReplaces(consumer)`.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const wrkExternalDepModule = "example.com/dep"

func setupWrkExternalConsumer(t *testing.T, rootDir string) (consumerTop, externalPath string) {
	t.Helper()
	consumerTop = filepath.Join(rootDir, "consumer")
	depMain := filepath.Join(rootDir, "dep")
	mkdirWrkExternal(t, consumerTop)
	mkdirWrkExternal(t, depMain)

	runWrkExternalGit(t, consumerTop, "init")
	runWrkExternalGit(t, depMain, "init")
	writeWrkExternalGoMod(t, depMain, "go.mod", "module "+wrkExternalDepModule+"\n\ngo 1.22\n")
	runWrkExternalGit(t, depMain, "add", "go.mod")
	runWrkExternalGit(t, depMain, "commit", "-m", "add dep module")

	externalPath = filepath.Join(consumerTop, "external", "mydep-main-2026-07-04")
	mkdirWrkExternal(t, filepath.Dir(externalPath))
	runWrkExternalGit(t, depMain, "worktree", "add", "-b", "mydep-main-2026-07-04", externalPath)
	return consumerTop, externalPath
}

func writeConsumerReplace(t *testing.T, consumerTop, externalPath string) {
	t.Helper()
	content := "module example.com/consumer\n\ngo 1.22\n\nrequire " + wrkExternalDepModule + " v0.0.0\n\nreplace " + wrkExternalDepModule + " => " + externalPath + "\n"
	writeWrkExternalGoMod(t, consumerTop, "go.mod", content)
}

func removeExternalWorktree(t *testing.T, depMain, externalPath string) {
	t.Helper()
	cmd := exec.Command("git", "-C", depMain, "worktree", "remove", "--force", externalPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree remove: %v\n%s", err, out)
	}
}

func mkdirWrkExternal(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeWrkExternalGoMod(t *testing.T, root, rel, content string) {
	t.Helper()
	if err := writeWrkExternalGoModFile(root, rel, content); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
}

func writeWrkExternalGoModFile(root, rel, content string) error {
	dir := filepath.Dir(filepath.Join(root, rel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, rel), []byte(content), 0o644)
}

func runWrkExternalGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func Setup(t *testing.T, req *Request) error {
	return nil
}

```