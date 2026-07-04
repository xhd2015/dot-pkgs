## Expected Output

```
would: wrk 0 deps
```

## Expected

- Exit code 0.
- Stdout is exactly `would: wrk 0 deps\n`.
- No `external/` directory; no `replace` directives; no `/external` in `.gitignore`.

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
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))
	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if len(mod.Replace) != 0 {
		t.Fatalf("go.mod should have no replace directives, got %+v", mod.Replace)
	}
	hasExternal, err := allDepsGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if hasExternal {
		t.Fatalf(".gitignore should have no /external line after dry-run")
	}
}
```