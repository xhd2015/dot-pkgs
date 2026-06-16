## Expected
- Exit code 0.
- Output contains "moved:".
- repo2/ no longer exists at its original location.
- wt/repo2/ exists with data.txt (repo2 was moved inside the worktree dir).
- wt/ still exists as a worktree with .git file.
- wt/.git still references repo1 (unchanged — repo2 is plain dir, no worktree
  update triggered).

## History
- repo1 chain: [repo1, wt(worktree)] (unchanged).
- repo2 chain: [repo2, wt/repo2] (new entry for the plain move).

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

	repo1 := filepath.Join(req.WorkRoot, "repo1")
	repo2 := filepath.Join(req.WorkRoot, "repo2")
	wt := filepath.Join(req.WorkRoot, "wt")
	repo2Inside := filepath.Join(wt, "repo2")

	assertContains(t, resp.Output, "moved:")

	// repo2 moved inside wt
	assertFileNotExists(t, repo2)
	assertFileExists(t, repo2Inside)
	assertFileExists(t, filepath.Join(repo2Inside, "data.txt"))

	// repo1 unchanged
	assertFileExists(t, repo1)
	assertFileExists(t, filepath.Join(repo1, "README.md"))

	// wt still exists as worktree, .git unchanged
	assertFileExists(t, wt)
	assertFileExists(t, filepath.Join(wt, ".git"))
	gitContent, err := os.ReadFile(filepath.Join(wt, ".git"))
	assertErrIsNil(t, err)
	assertContains(t, string(gitContent), repo1)

	// History: 2 projects
	h := readHistoryFile(t, req.ConfigHome)
	if h == nil {
		t.Fatal("expected history, got nil")
	}
	if len(h.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(h.Projects))
	}

	// repo1 chain: [repo1, wt(worktree)] (unchanged)
	proj, ok := h.Projects[repo1]
	if !ok {
		t.Fatalf("project key %s not found in history", repo1)
	}
	if len(proj.Locations) != 2 {
		t.Fatalf("expected 2 locations for %s, got %d", repo1, len(proj.Locations))
	}
	if proj.Locations[0].Path != repo1 {
		t.Fatalf("%s location[0]: expected %s, got %s", repo1, repo1, proj.Locations[0].Path)
	}
	if proj.Locations[1].Path != wt {
		t.Fatalf("%s location[1]: expected %s, got %s", repo1, wt, proj.Locations[1].Path)
	}

	// repo2 chain: [repo2, wt/repo2] (new entry for the plain move)
	proj2, ok2 := h.Projects[repo2]
	if !ok2 {
		t.Fatalf("project key %s not found in history", repo2)
	}
	if len(proj2.Locations) != 2 {
		t.Fatalf("expected 2 locations for %s, got %d", repo2, len(proj2.Locations))
	}
	if proj2.Locations[0].Path != repo2 {
		t.Fatalf("%s location[0]: expected %s, got %s", repo2, repo2, proj2.Locations[0].Path)
	}
	if proj2.Locations[1].Path != repo2Inside {
		t.Fatalf("%s location[1]: expected %s, got %s", repo2, repo2Inside, proj2.Locations[1].Path)
	}
}
```
