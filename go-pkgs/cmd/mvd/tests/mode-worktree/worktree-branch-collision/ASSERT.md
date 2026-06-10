## Expected
- Exit code 0.
- Output contains "worktree created:" but NOT "[branch: myfeature]".
- A branch starting with "myfeature-" was created instead.
- History records the worktree with the alternative branch name.

## Exit Code
- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "worktree created:")
	assertNotContains(t, resp.Output, "[branch: myfeature]")

	assertFileExists(t, filepath.Join(req.WorkRoot, "myfeature", ".git"))

	mainRepo := filepath.Join(req.WorkRoot, "main")
	wtDir := filepath.Join(req.WorkRoot, "myfeature")
	assertHistoryChain(t, req.ConfigHome, mainRepo, mainRepo, wtDir)

	h := readHistoryFile(t, req.ConfigHome)
	if h == nil {
		t.Fatal("expected history, got nil")
	}
	proj := h.Projects[mainRepo]
	branch := proj.Locations[1].Git.Branch
	if branch == "myfeature" {
		t.Fatal("branch name should not be 'myfeature' (collision)")
	}
	if !strings.HasPrefix(branch, "myfeature-") {
		t.Fatalf("branch name should start with 'myfeature-', got %q", branch)
	}
}
```
