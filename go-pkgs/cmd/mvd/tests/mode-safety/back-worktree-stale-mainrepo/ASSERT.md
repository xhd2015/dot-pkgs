## Expected
- Exit code 0 (worktree removed successfully).
- Output contains "worktree removed:".

## Behavior (After Fix)
- `resolveBackEntry` now allows --back on worktree entries at any position in the chain.
- `cmdWorktreeBackAt` removes the specific worktree entry while preserving
  subsequent plain moves.
- MainRepo is read from the worktree's .git file on disk, not from stale history.

## Side Effects
- repo/ no longer exists (was renamed to mid in step 2).
- mid/ exists (main repo's current location).
- wt/ no longer exists (worktree removed).

## History
- Chain: [repo, mid] (worktree entry removed, mid preserved).

## Exit Code
- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	repo := filepath.Join(req.WorkRoot, "repo")
	mid := filepath.Join(req.WorkRoot, "mid")
	wt := filepath.Join(req.WorkRoot, "wt")

	assertContains(t, resp.Output, "worktree removed:")

	// repo is gone, mid exists, wt is removed
	assertFileNotExists(t, repo)
	assertFileExists(t, mid)
	assertFileExists(t, filepath.Join(mid, "README.md"))
	assertFileExists(t, filepath.Join(mid, ".git"))
	assertFileNotExists(t, wt)

	// History: worktree entry removed, mid preserved
	assertHistoryChain(t, req.ConfigHome, repo,
		repo,
		mid,
	)
}
```
