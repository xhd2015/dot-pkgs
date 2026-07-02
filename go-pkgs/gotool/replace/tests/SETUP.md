# Scenario

**Feature**: shared library scans git repo for local replace directives

```
# Repo top -> Scan -> LocalFilesystemReplaces -> isIntraRepoReplace -> []ReplaceIssue
top -> scan go.mod files -> filter local replaces -> classify intra/extra repo -> return issues
```

## Preconditions

- A temporary git repository exists for each test case.
- The `gotool/replace` package is importable from this test tree.

## Steps

1. Initialize a temporary git repository for the test case.
2. Write go.mod files and fixture directories as specified by the leaf case.
3. Call `replace.CheckLocalReplaces(top)` with the repo top directory.

## Context

- `CheckLocalReplaces` scans all go.mod files under `top` using `scan.Scan`.
- Each module's `LocalFilesystemReplaces()` is called to find local-path replaces.
- Each local replace is classified as intra-repo or extra-repo via git toplevel comparison.
- The function returns all issues found; callers decide policy.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.RootDir = t.TempDir()
	if err := runGit(req.RootDir, "init"); err != nil {
		return err
	}
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) == 0 {
			return err
		}
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func writeGoMod(root string, rel string, content string) error {
	dir := filepath.Dir(filepath.Join(root, rel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, rel), []byte(content), 0o644)
}

func writeFile(root string, rel string, content string) error {
	dir := filepath.Dir(filepath.Join(root, rel))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, rel), []byte(content), 0o644)
}
```