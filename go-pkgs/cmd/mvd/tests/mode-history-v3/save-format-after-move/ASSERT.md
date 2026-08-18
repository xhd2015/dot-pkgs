## Expected
- Exit code 0 for the plain move.
- `history.json` has `version` `"3.0"`.
- Moves use `from`, `to`, `from_type`, `to_type` (no `prev`/`current`/`type`).
- Worktree move: `from_type=main`, `to_type=worktree`, `branch` set.
- Plain move after worktree: `from` is last non-worktree main (repo), `to_type=main`.

## Exit Code
- 0

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	repo := filepath.Join(req.WorkRoot, "repo")
	wt := filepath.Join(req.WorkRoot, "wt")
	dst := filepath.Join(req.WorkRoot, "dst")

	data, err := os.ReadFile(filepath.Join(req.ConfigHome, "history.json"))
	if err != nil {
		t.Fatalf("read history: %v", err)
	}
	raw := string(data)
	for _, legacy := range []string{`"prev"`, `"current"`, `"type"`} {
		if strings.Contains(raw, legacy) {
			t.Fatalf("history should not contain legacy field %s:\n%s", legacy, raw)
		}
	}

	var hf HistoryFile
	if err := json.Unmarshal(data, &hf); err != nil {
		t.Fatalf("parse history: %v", err)
	}
	if hf.Version != "3.0" {
		t.Fatalf("expected version 3.0, got %q", hf.Version)
	}

	proj, ok := hf.Projects[repo]
	if !ok {
		t.Fatalf("project %s not found in history", repo)
	}
	if len(proj.Moves) != 2 {
		t.Fatalf("expected 2 moves, got %d: %#v", len(proj.Moves), proj.Moves)
	}

	wtMove := proj.Moves[0]
	if wtMove.From != repo || wtMove.FromType != "main" || wtMove.To != wt || wtMove.ToType != "worktree" || wtMove.Branch == "" {
		t.Fatalf("unexpected worktree move: %#v", wtMove)
	}

	plainMove := proj.Moves[1]
	if plainMove.From != repo || plainMove.FromType != "main" || plainMove.To != dst || plainMove.ToType != "main" {
		t.Fatalf("unexpected plain move after worktree: %#v", plainMove)
	}

	assertHistoryChain(t, req.ConfigHome, repo, repo, wt, dst)
}
```