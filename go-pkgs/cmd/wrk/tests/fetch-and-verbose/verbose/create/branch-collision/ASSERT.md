## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WorkRoot}/wt` (fixed spawn path; branch ref reused).
- Stderr contains timestamp `worktree add --no-checkout` pre-command log line.
- Stderr contains git `worktree add` subprocess output for the add step (`Preparing worktree` or `HEAD is now at`).

## Exit Code

- 0

```go
import "path/filepath"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	wantPath := filepath.Join(req.WorkRoot, "wt")
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertStderrContainsGitSubcommand(t, resp.Stderr, "worktree add")
	assertContains(t, resp.Stderr, "--no-checkout")
	assertStderrVerboseFormat(t, resp.Stderr)
	assertStderrContainsWorktreeAddOutput(t, resp.Stderr)
}
```