## Expected

- Exit code 0; stderr empty.
- Stdout is exactly one line: `{abs my-repo}\tmain`.
- After scan, `LoadCacheEntry(--cache-dir, my-repo)` returns `ok=true` with
  `is_repo=true`, `repo_type=main`, non-empty `refreshed_at`, `scan_complete=true`.

## Side Effects

- Cold scan with `--cache-dir` populates the mirror under the caller-provided path
  (not the product default under `$HOME`).

## Exit Code

- `0`.

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}

	roots := rootsFromArgs(req.Args)
	if len(roots) != 1 {
		t.Fatalf("expected 1 --root, got %v", roots)
	}
	cacheDir := cacheDirFromArgs(req.Args)
	if cacheDir == "" {
		t.Fatal("expected --cache-dir in args")
	}

	repoPath := absPath(t, filepath.Join(roots[0], "my-repo"))
	wantGitDir := absPath(t, filepath.Join(repoPath, ".git"))
	wantLine := repoPath + "\tmain"
	got := strings.TrimSuffix(resp.Stdout, "\n")
	if got != wantLine {
		t.Fatalf("stdout = %q, want %q", got, wantLine)
	}

	entry, ok, loadErr := scan_repo.LoadCacheEntry(cacheDir, repoPath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry: %v", loadErr)
	}
	if !ok {
		t.Fatalf("expected cache entry for %s under --cache-dir %s", repoPath, cacheDir)
	}
	if !entry.IsRepo {
		t.Fatal("IsRepo = false, want true")
	}
	if entry.RepoType != string(scan_repo.RepoTypeMain) {
		t.Fatalf("RepoType = %q, want main", entry.RepoType)
	}
	if entry.GitDir != wantGitDir {
		t.Fatalf("GitDir = %q, want %q", entry.GitDir, wantGitDir)
	}
	if !entry.ScanComplete {
		t.Fatal("ScanComplete = false, want true")
	}
	if entry.RefreshedAt == "" {
		t.Fatal("RefreshedAt empty, want non-empty RFC3339")
	}
	if _, parseErr := time.Parse(time.RFC3339, entry.RefreshedAt); parseErr != nil {
		if _, parseErr2 := time.Parse(time.RFC3339Nano, entry.RefreshedAt); parseErr2 != nil {
			t.Fatalf("RefreshedAt %q not RFC3339: %v", entry.RefreshedAt, parseErr)
		}
	}
}
```
