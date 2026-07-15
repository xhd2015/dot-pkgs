## Expected

- `resp.ParseOK` is true.
- `resp.Parsed.Scope.PathPrefix` is `sub/`.
- `resp.Parsed.Version` is `0.2.3`.
- `resp.Parsed.Prerelease` is `beta`.
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
	if p.Scope.PathPrefix != "sub/" || p.Scope.VersionPrefix != "sub/v" {
		t.Fatalf("scope = %+v, want sub/ prefix", p.Scope)
	}
	if p.Version != "0.2.3" {
		t.Fatalf("Version = %q, want 0.2.3", p.Version)
	}
	if p.Prerelease != "beta" {
		t.Fatalf("Prerelease = %q, want beta", p.Prerelease)
	}
	if p.IsNumericRelease {
		t.Fatal("IsNumericRelease = true, want false")
	}
}
```