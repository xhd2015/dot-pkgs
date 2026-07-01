## Expected

- Non-zero exit code (sub-module local filesystem replace blocks `--done`).
- Stderr mentions `replace`.
- Stderr does NOT contain `no go.mod found` (the top-level go.mod exists; the
  block comes from the sub-module replace, not a missing go.mod).
- Consumer linked worktree still exists (merge-back did not run).

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit (sub-module replace guard), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertContains(t, resp.Stderr, "replace")
	assertNotContains(t, resp.Stderr, "no go.mod found")
	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}
```
