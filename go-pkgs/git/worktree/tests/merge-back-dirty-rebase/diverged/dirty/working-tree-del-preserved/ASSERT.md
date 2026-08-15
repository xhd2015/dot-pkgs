## Expected

- `MergeBack` returns no error, action `"rebased-and-merged"`.
- `delete-me.txt` is still absent from the source worktree (user's deletion preserved).
- `git status --porcelain` shows ` D delete-me.txt` (working-tree deleted, index has it), not `D `.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp.Action != "rebased-and-merged" {
		t.Fatalf("expected action 'rebased-and-merged', got %q", resp.Action)
	}

	// delete-me.txt must still be absent from working tree
	if _, err := os.Stat(filepath.Join(req.SourcePath, "delete-me.txt")); err == nil {
		t.Fatal("delete-me.txt was restored — user's deletion was lost")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat delete-me.txt: %v", err)
	}

	cmd := exec.Command("git", "-C", req.SourcePath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status: %v", err)
	}

	// Check porcelain: no staged changes
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		if len(line) < 2 {
			continue
		}
		indexStatus := line[0]
		if indexStatus != ' ' && indexStatus != '?' {
			t.Fatalf("unexpected staged change: %q\nfull status:\n%s", line, string(out))
		}
	}

	// Verify delete-me.txt shows as working-tree deletion ( D)
	hasWorkingTreeDel := false
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) >= 4 && line[0] == ' ' && line[1] == 'D' {
			pathPart := strings.TrimSpace(line[2:])
			if pathPart == "delete-me.txt" {
				hasWorkingTreeDel = true
			}
		}
	}
	if !hasWorkingTreeDel {
		t.Fatalf("expected  D delete-me.txt (working-tree deletion preserved) in git status, got:\n%s", string(out))
	}

	// feature merged into main
	sourceFeatCommit := branchCommit(t, req.MainRepo, "feature")
	mainHead := revParseHEAD(t, req.MainRepo)
	if !isAncestor(t, req.MainRepo, sourceFeatCommit, mainHead) {
		t.Fatal("feature branch commit should be ancestor of main HEAD after merge")
	}
}
```
