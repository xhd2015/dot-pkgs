## Expected
- Exit code 0; picker shows 4 entries with correct markers.
- dst shows `(external main)` when alive.
- Third move: `from=dst`, `from_type=main`, `to=wt2`, `to_type=worktree`.
- `readHistoryFile` round-trips moves into the expected location chain.

## Exit Code
- 0

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
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
	wt2 := filepath.Join(req.WorkRoot, "feature-b")

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 picker entries, got %d:\n%s", len(lines), resp.Output)
	}

	found := map[string]string{}
	for _, line := range lines {
		parts := strings.SplitN(line, " -> ", 2)
		if len(parts) == 2 {
			found[parts[1]] = line
		}
	}
	if !strings.Contains(found[dst], "(external main)") {
		t.Fatalf("dst line should contain (external main), got: %s", found[dst])
	}
	if strings.Contains(found[dst], "(dead external main)") {
		t.Fatalf("alive dst should not show (dead external main), got: %s", found[dst])
	}

	data, err := os.ReadFile(filepath.Join(req.ConfigHome, "history.json"))
	if err != nil {
		t.Fatalf("read history: %v", err)
	}
	var hf HistoryFile
	if err := json.Unmarshal(data, &hf); err != nil {
		t.Fatalf("parse history: %v", err)
	}
	if hf.Version != "3.0" {
		t.Fatalf("expected version 3.0, got %q", hf.Version)
	}
	proj := hf.Projects[root]
	if len(proj.Moves) != 3 {
		t.Fatalf("expected 3 moves, got %d: %#v", len(proj.Moves), proj.Moves)
	}
	ext := proj.Moves[2]
	if ext.From != dst || ext.FromType != "main" || ext.To != wt2 || ext.ToType != "worktree" || ext.Branch != "feature-b" {
		t.Fatalf("unexpected external-main worktree move: %#v", ext)
	}

	assertHistoryChain(t, req.ConfigHome, root, root, wt1, dst, wt2)
	assertHistoryWorktreeEntry(t, req.ConfigHome, root, 1, root, "feature-a")
	assertHistoryWorktreeEntry(t, req.ConfigHome, root, 3, dst, "feature-b")
}
```