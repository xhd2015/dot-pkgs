## Expected

- `Update` returns no error.
- Consumer requires `github.com/example/fixture-mod@v1.0.0`.
- Replace directive for the module is dropped.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Update(%q) failed: %v", req.TargetDir, resp.Err)
	}
	if resp.ModuleVersion != "v1.0.0" {
		t.Fatalf("require version = %q, want v1.0.0", resp.ModuleVersion)
	}
	if resp.HasReplace {
		t.Fatalf("replace for %q was not dropped", fixtureModulePath)
	}
}
```