## Expected

- Exit code 0.
- Stdout lists only dep2 at `./external/mydep2-main-2026-06-30`, then `wrk 1 deps`.
- The dep2 external path exists as a linked git worktree.
- Consumer `go.mod` has a `replace` for `example.com/dep2` at its external path.
- dep1's pre-existing `replace example.com/dep1 => ./external/preexisting` is unchanged.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantDep2 := allDepsExternalAbsPath(req.ConsumerTop, "mydep2")
	wantStdout := fmt.Sprintf("wrk example.com/dep2 at %s\nwrk 1 deps\n", allDepsExternalRelPath("mydep2"))
	if resp.Stdout != wantStdout {
		t.Fatalf("stdout mismatch:\nwant %q\n got %q", wantStdout, resp.Stdout)
	}

	assertFileExists(t, wantDep2)
	assertGitFileIsWorktreeLink(t, wantDep2)
	assertWorktreeListContains(t, req.ConsumerTop, wantDep2)

	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !allDepsHasReplaceForModule(mod, "example.com/dep2", wantDep2) {
		t.Fatalf("go.mod missing replace example.com/dep2 => %s: %+v", wantDep2, mod.Replace)
	}
	dep1Replace := allDepsReplacePathForModule(mod, "example.com/dep1")
	if dep1Replace != "./external/preexisting" {
		t.Fatalf("dep1 replace should be unchanged ./external/preexisting, got %q", dep1Replace)
	}
}
```
