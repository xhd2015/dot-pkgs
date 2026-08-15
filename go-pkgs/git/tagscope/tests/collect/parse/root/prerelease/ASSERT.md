## Expected

- `resp.ParseOK` is true.
- `resp.Parsed.FullName` is `v0.0.2-alpha`.
- `resp.Parsed.Scope.PathPrefix` is `""`.
- `resp.Parsed.Scope.VersionPrefix` is `v`.
- `resp.Parsed.Version` is `0.0.2`.
- `resp.Parsed.Prerelease` is `alpha`.
- `resp.Parsed.IsNumericRelease` is false.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	requireParseOK(t, resp.ParseOK)
	p := resp.Parsed
	if p.FullName != "v0.0.2-alpha" {
		t.Fatalf("FullName = %q, want v0.0.2-alpha", p.FullName)
	}
	if p.Scope.PathPrefix != "" || p.Scope.VersionPrefix != "v" {
		t.Fatalf("scope = %+v, want root v prefix", p.Scope)
	}
	if p.Version != "0.0.2" {
		t.Fatalf("Version = %q, want 0.0.2", p.Version)
	}
	if p.Prerelease != "alpha" {
		t.Fatalf("Prerelease = %q, want alpha", p.Prerelease)
	}
	if p.IsNumericRelease {
		t.Fatal("IsNumericRelease = true, want false")
	}
}
```