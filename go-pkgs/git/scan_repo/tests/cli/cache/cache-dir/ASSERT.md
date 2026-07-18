## Expected

- Exit code 0; stderr empty.
- Stdout is exactly one line: `{abs my-repo}\tmain`.
- After scan, `home/repos.json` under `--cache-dir` lists the main repo.
- Path `--cache-dir/mirror` does not exist.

## Side Effects

- Cold scan with `--cache-dir` seeds durable index under the caller-provided path
  (not the product default under `$HOME`). Dense mirror is retired.

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

	idx, ok, loadErr := scan_repo.LoadRepoIndex(cacheDir, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatalf("expected home/repos.json under --cache-dir %s", cacheDir)
	}
	found := false
	for _, e := range idx.Repos {
		if e.Path == repoPath && e.RepoType == string(scan_repo.RepoTypeMain) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("index missing main %s; entries=%v", repoPath, idx.Repos)
	}

	mirrorDir := filepath.Join(cacheDir, "mirror")
	if _, err := os.Stat(mirrorDir); err == nil {
		t.Fatalf("CLI cold scan created mirror at %s; dense mirror is retired", mirrorDir)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat mirror: %v", err)
	}
}
```
