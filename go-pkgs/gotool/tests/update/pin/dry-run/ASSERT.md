## Expected

- `Pin` returns no error (would succeed if applied).
- `PinResult` reports planned `ModulePath`, `Version=v1.0.0`, `Tag=v1.0.0`.
- On disk: require still `v0.0.1`, replace still present.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Pin(DryRun) failed: %v", resp.Err)
	}
	if resp.ModulePath != fixtureModulePath {
		t.Fatalf("ModulePath = %q, want %q", resp.ModulePath, fixtureModulePath)
	}
	if resp.ModuleVersion != "v1.0.0" {
		t.Fatalf("planned PinResult.Version = %q, want v1.0.0", resp.ModuleVersion)
	}
	if resp.Tag != "v1.0.0" {
		t.Fatalf("planned PinResult.Tag = %q, want v1.0.0", resp.Tag)
	}
	if resp.DiskVersion != "v0.0.1" {
		t.Fatalf("disk require version = %q, want v0.0.1 (unchanged)", resp.DiskVersion)
	}
	if !resp.HasReplace {
		t.Fatalf("replace for %q must remain on dry-run", fixtureModulePath)
	}
}
```
