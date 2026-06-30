## Expected

- Exit code 0.
- Stdout contains the main repo path and the linked worktree path.
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	linkedWT := filepath.Join(req.WorkRoot, "linked-wt")

	assertContains(t, resp.Stdout, mainRepo)
	assertContains(t, resp.Stdout, linkedWT)

	want := gitWorktreeList(t, mainRepo)
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch:\nwant:\n%q\ngot:\n%q", want, resp.Stdout)
	}

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```