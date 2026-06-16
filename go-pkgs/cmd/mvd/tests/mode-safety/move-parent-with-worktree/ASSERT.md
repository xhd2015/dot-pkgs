## Expected
- Exit code 0 (plain move succeeds).
- parent/ no longer exists physically.
- elsewhere/ exists, containing repo/ and wt/ (the whole tree moved).
- elsewhere/repo/ exists with README.md and .git dir.
- elsewhere/wt/ exists with .git file and README.md.

## Worktree .git is CORRECTLY UPDATED
- elsewhere/wt/.git references elsewhere/repo (the new location of the main repo).
  moveDir now recursively discovers git repos under the source directory and
  updates their linked worktrees' .git files after the rename.

## History
- The parent/repo history entry still exists with dead paths [parent/repo, parent/wt].
- A new entry for the parent dir itself was also created: [parent, elsewhere].

## Exit Code
- 0

```go
import (
	"os"
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
	elsewhereRepo := filepath.Join(elsewhere, "repo")
	elsewhereWt := filepath.Join(elsewhere, "wt")

	// Physical paths
	assertFileNotExists(t, parent)
	assertFileNotExists(t, parentRepo)
	assertFileNotExists(t, parentWt)

	assertFileExists(t, elsewhere)
	assertFileExists(t, elsewhereRepo)
	assertFileExists(t, filepath.Join(elsewhereRepo, "README.md"))
	assertFileExists(t, filepath.Join(elsewhereRepo, ".git"))

	assertFileExists(t, elsewhereWt)
	assertFileExists(t, filepath.Join(elsewhereWt, ".git"))
	assertFileExists(t, filepath.Join(elsewhereWt, "README.md"))

	// Worktree .git is correctly updated to reference elsewhere/repo
	gitContent, err := os.ReadFile(filepath.Join(elsewhereWt, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), elsewhereRepo)

	// History: 2 projects exist
	h := readHistoryFile(t, req.ConfigHome)
	if h == nil {
		t.Fatal("expected history, got nil")
	}
	if len(h.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(h.Projects))
	}

	// parent/repo entry still exists with dead paths
	proj, ok := h.Projects[parentRepo]
	if !ok {
		t.Fatalf("project key %s not found in history", parentRepo)
	}
	if len(proj.Locations) != 2 {
		t.Fatalf("expected 2 locations for %s, got %d", parentRepo, len(proj.Locations))
	}
	if proj.Locations[0].Path != parentRepo {
		t.Fatalf("%s location[0]: expected %s, got %s", parentRepo, parentRepo, proj.Locations[0].Path)
	}
	if proj.Locations[1].Path != parentWt {
		t.Fatalf("%s location[1]: expected %s, got %s", parentRepo, parentWt, proj.Locations[1].Path)
	}

	// parent dir history: new entry for the parent dir move
	proj2, ok2 := h.Projects[parent]
	if !ok2 {
		t.Fatalf("project key %s not found in history", parent)
	}
	if len(proj2.Locations) != 2 {
		t.Fatalf("expected 2 locations for %s, got %d", parent, len(proj2.Locations))
	}
	if proj2.Locations[0].Path != parent {
		t.Fatalf("%s location[0]: expected %s, got %s", parent, parent, proj2.Locations[0].Path)
	}
	if proj2.Locations[1].Path != elsewhere {
		t.Fatalf("%s location[1]: expected %s, got %s", parent, elsewhere, proj2.Locations[1].Path)
	}
}
```
