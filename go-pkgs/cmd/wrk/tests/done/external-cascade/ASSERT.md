## Expected

- Non-zero exit code (local filesystem replace remains after cascade).
- Stderr mentions local/filesystem replace must be resolved manually.
- External dependency worktree under `external/` no longer exists (cascade merge-back).
- Consumer linked worktree still exists (parent `--done` did not complete).

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit (local replace guard), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertContains(t, resp.Stderr, "replace")
	assertFileNotExists(t, req.ExternalWtDir)
	assertWorktreeListNotContains(t, req.MainRepo, req.ExternalWtDir)
	assertFileExists(t, req.WtDir)
}
```