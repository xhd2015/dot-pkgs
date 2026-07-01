## Expected Output

```
would: wrk 0 deps
```

## Expected

- Exit code 0.
- Stdout is exactly `would: wrk 0 deps\n`.
- `{consumerTop}/external/` does NOT exist.
- Consumer `go.mod` has no `replace` directives added.
- Consumer `.gitignore` has no `/external` line.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	if resp.Stdout != "would: wrk 0 deps\n" {
		t.Fatalf("stdout mismatch:\nwant %q\n got %q", "would: wrk 0 deps\n", resp.Stdout)
	}

	// external/ never created.
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))

	// go.mod unchanged: no replace directives.
	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if len(mod.Replace) != 0 {
		t.Fatalf("go.mod should have no replace directives, got %+v", mod.Replace)
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
