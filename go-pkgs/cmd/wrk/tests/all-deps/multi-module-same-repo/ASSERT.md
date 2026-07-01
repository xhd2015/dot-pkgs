## Expected

- Exit code 0.
- Stdout (exact) lists dep1 then dep2 (mod/scan Dir-sorted: `svc-a` < `svc-b`), each at its sub-dir, then `wrk 2 deps`:
  `wrk example.com/dep1 at ./external/myrepo-main-2026-06-30/svc-a`
  `wrk example.com/dep2 at ./external/myrepo-main-2026-06-30/svc-b`
  `wrk 2 deps`
- Exactly **one** external worktree at `{consumerTop}/external/myrepo-main-2026-06-30` (unsuffixed — both modules share it, no `-1` collision-suffixed worktree).
- Consumer `go.mod` has `replace example.com/dep1 => {abs}/external/myrepo-main-2026-06-30/svc-a` and `replace example.com/dep2 => {abs}/external/myrepo-main-2026-06-30/svc-b`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantRel1 := nestedExternalRelSubPath("myrepo", "svc-a")
	wantRel2 := nestedExternalRelSubPath("myrepo", "svc-b")
	wantStdout := fmt.Sprintf("wrk example.com/dep1 at %s\nwrk example.com/dep2 at %s\nwrk 2 deps\n", wantRel1, wantRel2)
	if resp.Stdout != wantStdout {
		t.Fatalf("stdout mismatch:\nwant %q\n got %q", wantStdout, resp.Stdout)
	}

	// Exactly one external worktree (dedup): the unsuffixed repo-root path.
	repoExternal := allDepsExternalAbsPath(req.ConsumerTop, "myrepo")
	assertFileExists(t, repoExternal)
	assertGitFileIsWorktreeLink(t, repoExternal)
	// The external dep worktree is owned by the DEP repo, not the consumer.
	assertWorktreeListContains(t, allDepsDepMainRepo(req.WorkRoot, "myrepo"), repoExternal)
	assertWorktreeListNotContains(t, req.ConsumerTop, repoExternal)
	// No collision-suffixed second worktree was created.
	suffixed := repoExternal + "-1"
	assertWorktreeListNotContains(t, req.ConsumerTop, suffixed)
	if _, err := os.Stat(suffixed); err == nil {
		t.Fatalf("no suffixed second worktree should exist at %s", suffixed)
	}

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	wantSub1 := nestedExternalAbsSubPath(req.ConsumerTop, "myrepo", "svc-a")
	wantSub2 := nestedExternalAbsSubPath(req.ConsumerTop, "myrepo", "svc-b")
	if !allDepsHasReplaceForModule(mod, "example.com/dep1", wantSub1) {
		t.Fatalf("go.mod missing replace example.com/dep1 => %s: %+v", wantSub1, mod.Replace)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep2", wantSub2) {
		t.Fatalf("go.mod missing replace example.com/dep2 => %s: %+v", wantSub2, mod.Replace)
	}
}
```
