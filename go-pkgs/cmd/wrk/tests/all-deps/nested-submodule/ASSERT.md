## Expected

- Exit code 0.
- Stdout (exact) is `wrk example.com/dep at ./external/myrepo-main-2026-06-30/services/dep` then `wrk 1 deps`.
- An external worktree for `myrepo` exists at `{consumerTop}/external/myrepo-main-2026-06-30` (the repo root) as a linked git worktree.
- Consumer `go.mod` has `replace example.com/dep => {abs}/external/myrepo-main-2026-06-30/services/dep` (the sub-module sub-directory, not the repo root).
- The sub-module `go.mod` is present inside the worktree at `external/myrepo-main-2026-06-30/services/dep/go.mod`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantRel := nestedExternalRelSubPath("myrepo", "services/dep")
	wantStdout := fmt.Sprintf("wrk example.com/dep at %s\nwrk 1 deps\n", wantRel)
	if resp.Stdout != wantStdout {
		t.Fatalf("stdout mismatch:\nwant %q\n got %q", wantStdout, resp.Stdout)
	}

	// One external worktree at the repo root.
	repoExternal := allDepsExternalAbsPath(req.ConsumerTop, "myrepo")
	assertFileExists(t, repoExternal)
	assertGitFileIsWorktreeLink(t, repoExternal)
	// The external dep worktree is owned by the DEP repo, not the consumer.
	assertWorktreeListContains(t, allDepsDepMainRepo(req.WorkRoot, "myrepo"), repoExternal)
	assertWorktreeListNotContains(t, req.ConsumerTop, repoExternal)

	// Replace points at the sub-module sub-directory, not the repo root.
	wantSubAbs := nestedExternalAbsSubPath(req.ConsumerTop, "myrepo", "services/dep")
	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep", wantSubAbs) {
		t.Fatalf("go.mod missing replace example.com/dep => %s: %+v", wantSubAbs, mod.Replace)
	}
	// The repo-root module (example.com/myrepo) must NOT have been replaced.
	if allDepsHasReplaceForModule(mod, "example.com/myrepo", "") {
		t.Fatalf("go.mod should not replace the non-required root module example.com/myrepo: %+v", mod.Replace)
	}

	// The sub-module go.mod is present inside the worktree.
	assertFileExists(t, filepath.Join(repoExternal, "services", "dep", "go.mod"))
}
```
