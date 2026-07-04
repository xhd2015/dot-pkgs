## Expected

- Exit code 0.
- One detailed status block with absolute `Dir`, `Compare with Remote:`, and `Worktrees: 0 Clean, 0 Dirty`.
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

	remote := compareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	block := projectStatusBlockTemplate(t, req.MainRepo, "clean", remote, "0 Clean, 0 Dirty")
	assert.Output(t, resp.Stdout, block)
}
```