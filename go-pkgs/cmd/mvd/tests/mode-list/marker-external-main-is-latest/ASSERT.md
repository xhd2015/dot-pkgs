## Expected
- Exit code 0.
- 3 entries (not 4): dst is both external main and latest, shown once.
- root: `(dead main)`
- wt1: `(dead worktree)`
- dst: `(external main)` — is also latest

## Exit Code
- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	root := filepath.Join(req.WorkRoot, "repo")
	wt1 := filepath.Join(req.WorkRoot, "feature-a")
	dst := filepath.Join(req.WorkRoot, "repo-moved")

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 picker entries (dst is external main + latest, no dup), got %d:\n%s", len(lines), resp.Output)
	}

	found := map[string]string{}
	for _, line := range lines {
		parts := strings.SplitN(line, " -> ", 2)
		if len(parts) == 2 {
			found[parts[1]] = line
		}
	}

	if found[root] == "" {
		t.Fatalf("root path %s not in picker output:\n%s", root, resp.Output)
	}
	if found[wt1] == "" {
		t.Fatalf("wt1 path %s not in picker output:\n%s", wt1, resp.Output)
	}
	if found[dst] == "" {
		t.Fatalf("dst path %s not in picker output:\n%s", dst, resp.Output)
	}

	if !strings.Contains(found[root], "(dead main)") {
		t.Fatalf("root line should contain (dead main), got: %s", found[root])
	}
	if !strings.Contains(found[wt1], "(dead worktree)") {
		t.Fatalf("wt1 line should contain (dead worktree), got: %s", found[wt1])
	}
	if !strings.Contains(found[dst], "(external main)") {
		t.Fatalf("dst line should contain (external main), got: %s", found[dst])
	}
}
```
