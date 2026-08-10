## Expected

- `ResolveLocalModules` returns no error.
- `IsDependency` is false (module path not referenced in consumer go.mod).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("ResolveLocalModules failed: %v", resp.Err)
	}
	if resp.IsDependency {
		t.Fatalf("IsDependency = true, want false for unknown module %q", depModulePath)
	}
}
```