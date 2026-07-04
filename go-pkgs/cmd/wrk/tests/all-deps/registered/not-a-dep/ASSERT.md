## Expected

- Exit code 0.
- Stdout is exactly `wrk 0 deps\n`.
- No `external/` directory; no new `replace` directives.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout != "wrk 0 deps\n" {
		t.Fatalf("stdout mismatch:\nwant %q\n got %q", "wrk 0 deps\n", resp.Stdout)
	}
	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if len(mod.Replace) != 0 {
		t.Fatalf("go.mod should have no replace directives, got %+v", mod.Replace)
	}
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))
}
```