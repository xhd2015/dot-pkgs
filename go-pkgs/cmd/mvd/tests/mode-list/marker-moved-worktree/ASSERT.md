## Expected
- Exit code 0.
- 3 picker entries: repo, wt, moved-wt.
- repo: `(main)` — root in a project with worktrees.
- moved-wt: should be `(worktree)` but currently shows `(external main)` (BUG).

## BUG
After `mvd wt moved-wt`, the moved worktree's LocationEntry is created
without Git metadata (`Git: nil`). The picker computes `isWt=false` for
moved-wt because `loc.Git.Type != "worktree"`, even though the path IS
a git worktree on disk. With `hasWorktree=true` (from the still-present
wt entry), moved-wt gets classified as `(external main)`.

Expected: `(worktree)` — because moved-wt IS a worktree on disk.

## Exit Code
- 0

```go
import (
	"fmt"
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	repo := filepath.Join(req.WorkRoot, "repo")
	wt := filepath.Join(req.WorkRoot, "wt")
	movedWt := filepath.Join(req.WorkRoot, "moved-wt")

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

	if found[repo] == "" {
		t.Fatalf("repo path %s not in picker output:\n%s", repo, resp.Output)
	}
	if found[movedWt] == "" {
		t.Fatalf("moved-wt path %s not in picker output:\n%s", movedWt, resp.Output)
	}

	// repo should be (main)
	if !strings.Contains(found[repo], "(main)") {
		t.Fatalf("repo should have (main) marker, got: %s", found[repo])
	}

	// BUG REPRODUCTION:
	// moved-wt IS a worktree on disk but shows (external main).
	// This assertion will FAIL until the bug is fixed.
	if strings.Contains(found[movedWt], "(external main)") {
		msg := fmt.Sprintf(
			"BUG: moved-wt shows (external main) instead of (worktree)\n"+
				"  moved-wt IS a git worktree on disk (contains .git file)\n"+
				"  but cmdMove did not preserve the worktree Git metadata.\n"+
				"  Picker line: %s\n"+
				"  Full output:\n%s",
			found[movedWt], resp.Output,
		)
		t.Fatal(msg)
	}
	if !strings.Contains(found[movedWt], "(worktree)") {
		t.Fatalf("BUG: moved-wt should have (worktree) marker, got: %s\nFull output:\n%s", found[movedWt], resp.Output)
	}

	// Also verify history has correct worktree metadata on moved-wt.
	// The second move entry (index 2) should have GitType="worktree".
	assertHistoryWorktreeEntry(t, req.ConfigHome, repo, 2, repo, "wt")

	// wt should no longer exist on disk (it was moved)
	// It appears as (dead worktree) in the picker
	if found[wt] != "" && strings.Contains(found[wt], "(dead worktree)") {
		// Acceptable: old wt path is dead on disk, marker is correct
	}
}
```
