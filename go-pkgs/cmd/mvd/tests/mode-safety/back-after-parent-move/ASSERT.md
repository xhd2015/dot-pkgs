## Expected
- Exit code 0 (--back succeeds).
- elsewhere/ no longer exists.
- parent/ exists again with repo/ and wt/ inside.
- parent/repo/ exists with README.md and .git dir.
- parent/wt/ exists with .git file and README.md.

## Sub-project history cleanup (Fixed)
- After --back restores the parent directory, orphaned sub-project entries
  (like parent/repo) are automatically cleaned up from history.
- Only the parent entry remains: [parent] (back removed the elsewhere step).

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

	parent := filepath.Join(req.WorkRoot, "parent")
	parentRepo := filepath.Join(parent, "repo")
	parentWt := filepath.Join(parent, "wt")
	elsewhere := filepath.Join(req.WorkRoot, "elsewhere")

	assertContains(t, resp.Output, "moved back:")

	// elsewhere should be gone, parent restored
	assertFileNotExists(t, elsewhere)
	assertFileExists(t, parent)
	assertFileExists(t, parentRepo)
	assertFileExists(t, filepath.Join(parentRepo, "README.md"))
	assertFileExists(t, filepath.Join(parentRepo, ".git"))
	assertFileExists(t, parentWt)
	assertFileExists(t, filepath.Join(parentWt, ".git"))
	assertFileExists(t, filepath.Join(parentWt, "README.md"))

	// History: only 1 project (parent) — sub-project cleaned up
	h := readHistoryFile(t, req.ConfigHome)
	if h == nil {
		t.Fatal("expected history, got nil")
	}
	if len(h.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(h.Projects))
	}

	// parent dir history: after back, just [parent] (elsewhere step removed)
	proj, ok := h.Projects[parent]
	if !ok {
		t.Fatalf("project key %s not found in history", parent)
	}
	if len(proj.Locations) != 1 {
		t.Fatalf("expected 1 location for %s, got %d", parent, len(proj.Locations))
	}
	if proj.Locations[0].Path != parent {
		t.Fatalf("%s location[0]: expected %s, got %s", parent, parent, proj.Locations[0].Path)
	}
}
```
