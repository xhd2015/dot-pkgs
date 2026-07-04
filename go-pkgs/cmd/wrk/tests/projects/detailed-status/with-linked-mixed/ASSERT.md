## Expected

- Exit code 0.
- `Worktrees: 2 Clean, 1 Dirty` in the project block.
- Stderr is empty.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertProjectsBlocksSeparated(t, resp.Stdout, 1)

	wantSummary := linkedWorktreeSummary(t, req.MainRepo)
	if wantSummary != "2 Clean, 1 Dirty" {
		t.Fatalf("helper summary: want 2 Clean, 1 Dirty, got %q", wantSummary)
	}
	remote := compareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	block := projectStatusBlockTemplate(t, req.MainRepo, "clean", remote, wantSummary)
	assert.Output(t, resp.Stdout, block)
}
```