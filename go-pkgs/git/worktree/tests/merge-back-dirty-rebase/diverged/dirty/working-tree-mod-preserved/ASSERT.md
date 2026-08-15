## Expected

- `MergeBack` returns no error, action `"rebased-and-merged"`.
- `README.md` in the source worktree still contains `# user modified` (user's content preserved).
- `git status --porcelain` shows ` M README.md` (working-tree modified, index clean), not `M `.

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

	// User's modification to README.md must be preserved
	content, err := os.ReadFile(filepath.Join(req.SourcePath, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	if string(content) != "# user modified\n" {
		t.Fatalf("README.md content was overwritten, expected '# user modified\\n', got %q", string(content))
	}

	// Check porcelain: no staged changes, only working-tree dirty
	cmd := exec.Command("git", "-C", req.SourcePath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status: %v", err)
	}
	hasWorkingTreeMod := false
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
		// porcelain: " XY path" where X=index, Y=working-tree.
		// Field split skips the leading space and separates status from path.
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "README.md" {
			if indexStatus == ' ' && len(line) > 1 && line[1] == 'M' {
				hasWorkingTreeMod = true
			}
		}
	}
	if !hasWorkingTreeMod {
		t.Fatalf("expected working-tree modification of README.md in git status, got:\n%s", string(out))
	}

	// feature merged into main
	sourceFeatCommit := branchCommit(t, req.MainRepo, "feature")
	mainHead := revParseHEAD(t, req.MainRepo)
	if !isAncestor(t, req.MainRepo, sourceFeatCommit, mainHead) {
		t.Fatal("feature branch commit should be ancestor of main HEAD after merge")
	}
}
```
