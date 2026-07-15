## Expected

- `resp.ParseOK` is true.
- `resp.Parsed.FullName` is `sub/v0.2.3`.
- `resp.Parsed.Scope.PathPrefix` is `sub/`.
- `resp.Parsed.Scope.VersionPrefix` is `sub/v`.
- `resp.Parsed.Version` is `0.2.3`.
- `resp.Parsed.Prerelease` is empty.
- `resp.Parsed.IsNumericRelease` is true.

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
	if p.FullName != "sub/v0.2.3" {
		t.Fatalf("FullName = %q, want sub/v0.2.3", p.FullName)
	}
	if p.Scope.PathPrefix != "sub/" {
		t.Fatalf("PathPrefix = %q, want sub/", p.Scope.PathPrefix)
	}
	if p.Scope.VersionPrefix != "sub/v" {
		t.Fatalf("VersionPrefix = %q, want sub/v", p.Scope.VersionPrefix)
	}
	if p.Version != "0.2.3" || p.Prerelease != "" || !p.IsNumericRelease {
		t.Fatalf("parsed = %+v, want 0.2.3 numeric release", p)
	}
}
```