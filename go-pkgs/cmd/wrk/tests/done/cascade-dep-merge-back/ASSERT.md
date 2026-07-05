## Expected

- Non-zero exit code (the consumer still carries `replace => ./external/...`, so
  the local-replace guard blocks the consumer's own merge-back after the cascade
  runs).
- Stderr mentions `replace` (guard message).
- The dep fix committed on the external worktree was merged back into the dep
  repo's main — `git -C <depRepo> log` contains `dep fix on worktree`. This is
  the merge-back guarantee: ahead dep work is not discarded.
- External dependency worktree under `external/` no longer exists.
- Consumer linked worktree still exists (guard blocked the consumer merge-back).

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit (guard blocks consumer after cascade), got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	assertContains(t, resp.Stderr, "replace")

	// Merge-back: the dep fix landed in the dep repo's main.
	depLog := gitOutputIsolated(t, req.DepPath, "log", "--oneline")
	assertContains(t, depLog, "dep fix on worktree")

	// External dep worktree removed by the cascade.
	assertFileNotExists(t, req.ExternalWtDir)
	assertWorktreeListNotContains(t, req.DepPath, req.ExternalWtDir)

	// Consumer worktree remains (guard blocked its merge-back).
	assertFileExists(t, req.WtDir)
}
```
