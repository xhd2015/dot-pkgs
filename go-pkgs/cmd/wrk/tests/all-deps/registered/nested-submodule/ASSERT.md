## Expected

- Exit code 0.
- Stdout: `wrk example.com/dep at ./external/myrepo-main-2026-06-30/services/dep` then `wrk 1 deps`.
- One external worktree at repo root; replace points at the sub-module directory.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	myrepo := allDepsDepDir(req.WorkRoot, "myrepo")
	wantRel := nestedExternalRelSubPath("myrepo", "services/dep")
	wantStdout := fmt.Sprintf("wrk example.com/dep at %s\nwrk 1 deps\n", wantRel)
	if resp.Stdout != wantStdout {
		t.Fatalf("stdout mismatch:\nwant %q\n got %q", wantStdout, resp.Stdout)
	}

	repoExternal := allDepsExternalAbsPath(req.ConsumerTop, "myrepo")
	assertFileExists(t, repoExternal)
	assertGitFileIsWorktreeLink(t, repoExternal)
	assertWorktreeListContains(t, allDepsDepMainRepo(myrepo), repoExternal)
	assertWorktreeListNotContains(t, req.ConsumerTop, repoExternal)

	wantSubAbs := nestedExternalAbsSubPath(req.ConsumerTop, "myrepo", "services/dep")
	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep", wantSubAbs) {
		t.Fatalf("go.mod missing replace example.com/dep => %s: %+v", wantSubAbs, mod.Replace)
	}
	if allDepsHasReplaceForModule(mod, "example.com/myrepo", "") {
		t.Fatalf("go.mod should not replace non-required root module: %+v", mod.Replace)
	}
	assertFileExists(t, filepath.Join(repoExternal, "services", "dep", "go.mod"))
}
```