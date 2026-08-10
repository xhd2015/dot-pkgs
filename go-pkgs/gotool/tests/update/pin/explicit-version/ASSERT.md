## Expected

- `Pin` returns no error.
- `PinResult.Version` is the forced `v0.0.5` (not latest tag `v1.0.0`).
- Consumer require on disk is `v0.0.5`.
- Replace for the module is dropped.
- `Tag` may be empty when Version is forced without tag lookup (do not require a specific Tag).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Pin(Version=%q) failed: %v", req.Version, resp.Err)
	}
	if resp.ModulePath != fixtureModulePath {
		t.Fatalf("ModulePath = %q, want %q", resp.ModulePath, fixtureModulePath)
	}
	if resp.ModuleVersion != "v0.0.5" {
		t.Fatalf("PinResult.Version = %q, want v0.0.5", resp.ModuleVersion)
	}
	if resp.DiskVersion != "v0.0.5" {
		t.Fatalf("disk require version = %q, want v0.0.5", resp.DiskVersion)
	}
	if resp.HasReplace {
		t.Fatalf("replace for %q was not dropped", fixtureModulePath)
	}
}
```
