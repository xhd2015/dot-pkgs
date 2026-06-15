## Expected
- Exit code 0.
- 3 entries: root with `(main)` marker, both worktrees with `(worktree)` marker.
- All three full paths appear.

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
	wt2 := filepath.Join(req.WorkRoot, "feature-b")

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 picker entries, got %d:\n%s", len(lines), resp.Output)
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
		t.Fatalf("worktree %s not in picker output:\n%s", wt1, resp.Output)
	}
	if found[wt2] == "" {
		t.Fatalf("worktree %s not in picker output:\n%s", wt2, resp.Output)
	}

	if !strings.Contains(found[root], "(main)") {
		t.Fatalf("root line should contain (main) marker, got: %s", found[root])
	}
	if !strings.Contains(found[wt1], "(worktree)") {
		t.Fatalf("wt1 line should contain (worktree) marker, got: %s", found[wt1])
	}
	if !strings.Contains(found[wt2], "(worktree)") {
		t.Fatalf("wt2 line should contain (worktree) marker, got: %s", found[wt2])
	}
}
```
