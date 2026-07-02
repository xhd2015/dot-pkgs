## Expected

- `MergeBack` returns no error, action is `"rebased-and-merged"`.
- `git status --porcelain` in the source worktree shows NO staged changes (first-column entries like `M `, `D `, `A `).
- Only untreated (second-column) entries from the intentional dirtiness (`?? dirty.txt`).

## Expected Output

After merge-back, `git status --porcelain` should only show:
```
?? dirty.txt
```

No lines starting with `M`, `D`, or `A` in the first column (staged index changes).

```go
import (
	"os/exec"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action 'rebased-and-merged', got %q", resp.Action)
	}

	// Check git status in source worktree.
	// Do NOT trim lines — porcelain uses leading XY columns:
	//   "XY path" where X=index, Y=working-tree.
	// Trimming turns " M" (working-tree modified, index clean)
	// into "M" (index modified) — a false positive.
	cmd := exec.Command("git", "-C", req.SourcePath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		if len(line) < 2 {
			continue
		}
		indexStatus := line[0]
		if indexStatus != ' ' && indexStatus != '?' {
			t.Fatalf("unexpected staged change in source worktree after merge-back: %q\nfull status:\n%s", line, string(out))
		}
	}

	// source worktree still exists
	if !hasDir(req.SourcePath) {
		t.Fatal("source worktree should still exist")
	}

	// feature was merged into main
	sourceFeatCommit := branchCommit(t, req.MainRepo, "feature")
	mainHead := revParseHEAD(t, req.MainRepo)
	if !isAncestor(t, req.MainRepo, sourceFeatCommit, mainHead) {
		t.Fatal("feature branch commit should be ancestor of main HEAD after merge")
	}
}
```
