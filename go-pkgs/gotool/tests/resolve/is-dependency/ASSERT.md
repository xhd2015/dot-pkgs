## Expected

- `ResolveLocalModules` returns no error.
- `IsDependency` is true for `example.com/dep`.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("ResolveLocalModules failed: %v", resp.Err)
	}
	if !resp.IsDependency {
		t.Fatalf("IsDependency = false, want true for %q", depModulePath)
	}
	if resp.ModulePath != depModulePath {
		t.Fatalf("module path = %q, want %q", resp.ModulePath, depModulePath)
	}
}
```