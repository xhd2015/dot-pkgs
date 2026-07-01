## Expected Output

```
would: wrk example.com/dep1 at ./external/mydep1-main-2026-06-30
would: wrk example.com/dep2 at ./external/mydep2-main-2026-06-30
would: wrk 2 deps
```

## Expected

- Exit code 0.
- Stdout is exactly the three `would:` lines above (dep1 then dep2, scan path-sorted order), each at `./external/mydepN-main-2026-06-30`, then a final `would: wrk 2 deps`.
- `{consumerTop}/external/` does NOT exist (never created).
- Consumer `go.mod` is unchanged: NO `replace` for `example.com/dep1` and NO `replace` for `example.com/dep2` (replace list empty or untouched).
- Consumer `.gitignore` has NO `/external` line.
- No git worktree created under `external/` for either dep.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantStdout := fmt.Sprintf("would: wrk example.com/dep1 at %s\nwould: wrk example.com/dep2 at %s\nwould: wrk 2 deps\n",
		allDepsExternalRelPath("mydep1"), allDepsExternalRelPath("mydep2"))
	if resp.Stdout != wantStdout {
		t.Fatalf("stdout mismatch:\nwant %q\n got %q", wantStdout, resp.Stdout)
	}

	// Core dry-run guarantee: external/ never created.
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))

	// No git worktree created for either dep under external/.
	wantDep1 := allDepsExternalAbsPath(req.ConsumerTop, "mydep1")
	wantDep2 := allDepsExternalAbsPath(req.ConsumerTop, "mydep2")
	assertFileNotExists(t, wantDep1)
	assertFileNotExists(t, wantDep2)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantDep1)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantDep2)

	// go.mod unchanged: no replace for either dep.
	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if allDepsHasReplaceForModule(mod, "example.com/dep1", "") {
		t.Fatalf("go.mod should have no replace for example.com/dep1 after dry-run, got %+v", mod.Replace)
	}
	if allDepsHasReplaceForModule(mod, "example.com/dep2", "") {
		t.Fatalf("go.mod should have no replace for example.com/dep2 after dry-run, got %+v", mod.Replace)
	}

	// .gitignore unchanged: no /external line.
	hasExternal, err := allDepsGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if hasExternal {
		t.Fatalf(".gitignore should have no /external line after dry-run")
	}
}
```
