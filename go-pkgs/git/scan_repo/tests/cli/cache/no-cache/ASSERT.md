## Expected

- Exit 0; stdout lists the repo.
- No `home/repos.json`, no `walk.jsonl`, no `mirror/` under `--cache-dir`.

## Exit Code

- `0`.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	roots := rootsFromArgs(req.Args)
	cacheDir := cacheDirFromArgs(req.Args)
	if cacheDir == "" {
		t.Fatal("expected --cache-dir")
	}
	repoPath := absPath(t, filepath.Join(roots[0], "my-repo"))
	if !strings.Contains(resp.Stdout, repoPath) {
		t.Fatalf("stdout missing %s:\n%s", repoPath, resp.Stdout)
	}

	for _, rel := range []string{
		filepath.Join("home", "repos.json"),
		filepath.Join("home", "walk.jsonl"),
		"mirror",
	} {
		p := filepath.Join(cacheDir, rel)
		if _, err := os.Stat(p); err == nil {
			t.Fatalf("--no-cache but found %s", p)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", p, err)
		}
	}
}
```
