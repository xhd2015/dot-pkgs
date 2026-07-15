## Expected

- Exit code 0; stderr empty.
- Stdout is exactly one line: `{abs my-repo}\tmain`.
- No `entry.json` under `--cache-dir` (walk finds none).
- `LoadCacheEntry(cacheDir, my-repo)` returns `ok=false`.

## Side Effects

- Successful discovery does not populate the cache store when `--no-cache` is set.

## Exit Code

- `0`.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	wantLine := repoPath + "\tmain"
	got := strings.TrimSuffix(resp.Stdout, "\n")
	if got != wantLine {
		t.Fatalf("stdout = %q, want %q", got, wantLine)
	}

	entry, ok, loadErr := scan_repo.LoadCacheEntry(cacheDir, repoPath)
	if loadErr != nil {
		t.Fatalf("LoadCacheEntry: %v", loadErr)
	}
	if ok {
		t.Fatalf("--no-cache should not write cache entry, got %+v", entry)
	}

	var found []string
	_ = filepath.WalkDir(cacheDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && d.Name() == "entry.json" {
			found = append(found, path)
		}
		return nil
	})
	if len(found) != 0 {
		t.Fatalf("--no-cache but found entry.json under cache-dir: %v", found)
	}
}
```
