## Expected

- `Replace` returns no error.
- Consumer `go.mod` contains `replace example.com/dep => <abs depDir>`.

## Exit Code

- N/A (library call)

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Replace(%q) failed: %v", req.TargetDir, resp.Err)
	}
	if resp.ModulePath != depModulePath {
		t.Fatalf("module path = %q, want %q", resp.ModulePath, depModulePath)
	}
	if !resp.HasReplace {
		t.Fatalf("expected replace directive for %s -> %s", depModulePath, resp.AbsDir)
	}
}
```