## Expected

- `Pin` returns no error.
- `PinResult.ModulePath` is `github.com/example/fixture-mod`.
- `PinResult.Version` is `v1.0.0` (latest tag stripped to require version).
- `PinResult.Tag` is `v1.0.0`.
- Consumer require on disk is `v1.0.0`.
- Replace directive for the module is dropped.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Pin(ConsumerDir=%q, DepDir=%q) failed: %v", req.ConsumerDir, req.TargetDir, resp.Err)
	}
	if resp.ModulePath != fixtureModulePath {
		t.Fatalf("ModulePath = %q, want %q", resp.ModulePath, fixtureModulePath)
	}
	if resp.ModuleVersion != "v1.0.0" {
		t.Fatalf("PinResult.Version = %q, want v1.0.0", resp.ModuleVersion)
	}
	if resp.Tag != "v1.0.0" {
		t.Fatalf("PinResult.Tag = %q, want v1.0.0", resp.Tag)
	}
	if resp.DiskVersion != "v1.0.0" {
		t.Fatalf("disk require version = %q, want v1.0.0", resp.DiskVersion)
	}
	if resp.HasReplace {
		t.Fatalf("replace for %q was not dropped", fixtureModulePath)
	}
}
```
