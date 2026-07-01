## Expected

- Exit code 0.
- Stdout lists dep1 at `./external/mydep1-main-2026-06-30`, then `wrk 1 deps`.
- The dep1 external path exists as a linked git worktree.
- Consumer `go.mod` has a `replace` for `example.com/dep1` but NOT for `example.com/dep2`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantDep1 := allDepsExternalAbsPath(req.ConsumerTop, "mydep1")
	wantStdout := fmt.Sprintf("wrk example.com/dep1 at %s\nwrk 1 deps\n", allDepsExternalRelPath("mydep1"))
	if resp.Stdout != wantStdout {
		t.Fatalf("stdout mismatch:\nwant %q\n got %q", wantStdout, resp.Stdout)
	}

	assertFileExists(t, wantDep1)
	assertGitFileIsWorktreeLink(t, wantDep1)
	// The external dep worktree is owned by its DEP repo, not the consumer.
	assertWorktreeListContains(t, allDepsDepMainRepo(req.WorkRoot, "mydep1"), wantDep1)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantDep1)

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep1", wantDep1) {
		t.Fatalf("go.mod missing replace example.com/dep1 => %s: %+v", wantDep1, mod.Replace)
	}
	if allDepsHasReplaceForModule(mod, "example.com/dep2", "") {
		t.Fatalf("go.mod should NOT replace example.com/dep2: %+v", mod.Replace)
	}
}
```
