## Expected

- `resp.ParseOK` is true.
- `resp.Parsed.FullName` is `pkg/api/v1.0.0-dev`.
- `resp.Parsed.Scope.PathPrefix` is `pkg/api/`.
- `resp.Parsed.Scope.VersionPrefix` is `pkg/api/v`.
- `resp.Parsed.Version` is `1.0.0`.
- `resp.Parsed.Prerelease` is `dev`.
- `resp.Parsed.IsNumericRelease` is false.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	requireParseOK(t, resp.ParseOK)
	p := resp.Parsed
	if p.FullName != "pkg/api/v1.0.0-dev" {
		t.Fatalf("FullName = %q, want pkg/api/v1.0.0-dev", p.FullName)
	}
	if p.Scope.PathPrefix != "pkg/api/" {
		t.Fatalf("PathPrefix = %q, want pkg/api/", p.Scope.PathPrefix)
	}
	if p.Scope.VersionPrefix != "pkg/api/v" {
		t.Fatalf("VersionPrefix = %q, want pkg/api/v", p.Scope.VersionPrefix)
	}
	if p.Version != "1.0.0" || p.Prerelease != "dev" || p.IsNumericRelease {
		t.Fatalf("parsed = %+v, want 1.0.0-dev prerelease", p)
	}
}
```