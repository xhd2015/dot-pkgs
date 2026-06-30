## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-feature-foo-2026-06-30` (slash → `-` in path token).
- Git branch name remains `feature/foo-2026-06-30` (slash preserved, date appended).
- Worktree directory exists with linked `.git` file.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	wantPath := worktreePath(req.WrkHome, "myrepo", "feature-foo", wrkDate, 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchExists(t, req.RepoDir, branchName("feature/foo", wrkDate, 0))
	assertBranchCheckedOutInWorktree(t, wantPath, branchName("feature/foo", wrkDate, 0))
	assertWorktreeListContains(t, req.RepoDir, wantPath)
}
```