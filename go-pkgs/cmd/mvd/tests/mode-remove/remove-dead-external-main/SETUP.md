# Scenario

Remove a dead external main path from a chain via `findEntry`, not only `hist[absDir]`.

Reproduces the reported bug: picker shows `(dead external main)` but `mvd --rm` returned `hint: no recorded entry`.

Chain: root → wt1 → dst (dead external main) → wt2

## Steps
- Seed history with the 4-location chain (same topology as `marker-dead-external-main`).
- Create root, wt1, wt2 directories but NOT dst (dst is dead on disk).
- Run `--picker-list` as sanity: dst must show `(dead external main)`.
- Run `mvd --rm <dst>` using the absolute path.

```go
import (
	"fmt"
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	root := filepath.Join(req.WorkRoot, "repo")
	wt1 := filepath.Join(req.WorkRoot, "feature-a")
	dst := filepath.Join(req.WorkRoot, "repo-moved")
	wt2 := filepath.Join(req.WorkRoot, "feature-b")

	mkdirAll(t, root)
	mkdirAll(t, wt1)
	mkdirAll(t, wt2)

	hf := HistoryFile{
		Version: "3.0",
		Projects: map[string]ProjectEntry{
			root: {
				Root: root,
				Moves: []MoveEntry{
					{From: root, FromType: "main", To: wt1, ToType: "worktree", Branch: "feature-a"},
					{From: root, FromType: "main", To: dst, ToType: "main"},
					{From: dst, FromType: "main", To: wt2, ToType: "worktree", Branch: "feature-b"},
				},
			},
		},
	}
	writeHistoryFile(t, req.ConfigHome, hf)

	req.Args = []string{"--picker-list"}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("picker-list: %s", resp.Output)
	}
	if !strings.Contains(resp.Output, "(dead external main)") {
		return fmt.Errorf("picker sanity: expected (dead external main), got:\n%s", resp.Output)
	}

	req.Args = []string{"--rm", dst}
	return nil
}
```