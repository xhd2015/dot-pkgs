## Expected

- Scan succeeds with exactly one discovered main repo (`my-repo`).
- No `home/repos.json`, no `home/walk.jsonl`, no `mirror/` under CacheRoot.

## Errors

- `err` is nil.

## Side Effects

- CacheRoot remains free of durable cache artifacts despite successful Scan.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(resp.Repos))
	}
	repoPath := absPath(t, filepath.Join(req.Roots[0], "my-repo"))
	if resp.Repos[0].Path != repoPath {
		t.Fatalf("Path = %q, want %q", resp.Repos[0].Path, repoPath)
	}

	for _, rel := range []string{
		filepath.Join("home", "repos.json"),
		filepath.Join("home", "walk.jsonl"),
		filepath.Join("home", "walk.cursor.json"),
		"mirror",
	} {
		p := filepath.Join(req.CacheRoot, rel)
		if _, err := os.Stat(p); err == nil {
			t.Fatalf("NoCache=true but found cache artifact %s", p)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", p, err)
		}
	}
}
```
