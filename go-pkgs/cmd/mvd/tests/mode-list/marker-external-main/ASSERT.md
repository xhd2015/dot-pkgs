## Expected
- Exit code 0.
- 4 entries shown (the bug fix ensures dst is included).
- root: `(dead main)` — dead + isRoot + isMain
- wt1: `(dead worktree)` — dead + isWorktree
- dst: `(external main)` — alive, isExternalMain (isMain && !isRoot)
- wt2: `(worktree)` — alive, isWorktree, also latest
- dst MUST be present (this is the regression being fixed).

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	root := filepath.Join(req.WorkRoot, "repo")
	wt1 := filepath.Join(req.WorkRoot, "feature-a")
	dst := filepath.Join(req.WorkRoot, "repo-moved")
	wt2 := filepath.Join(req.WorkRoot, "feature-b")

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 picker entries (including external main), got %d:\n%s", len(lines), resp.Output)
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
		t.Fatalf("dst (external main) path %s not in picker output (this is the bug regression):\n%s", dst, resp.Output)
	}
	if found[wt2] == "" {
		t.Fatalf("wt2 path %s not in picker output:\n%s", wt2, resp.Output)
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
	// Ensure dst does NOT incorrectly have (worktree) marker
	if strings.Contains(found[dst], "(worktree)") {
		t.Fatalf("dst line should NOT contain (worktree), got: %s", found[dst])
	}
	if !strings.Contains(found[wt2], "(worktree)") {
		t.Fatalf("wt2 line should contain (worktree), got: %s", found[wt2])
	}
}
```
