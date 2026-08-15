## Expected

- `resp.ParseOK` is true.
- `resp.Parsed.FullName` is `v0.0.1`.
- `resp.Parsed.Scope.PathPrefix` is `""`.
- `resp.Parsed.Scope.VersionPrefix` is `v`.
- `resp.Parsed.Version` is `0.0.1`.
- `resp.Parsed.Prerelease` is empty.
- `resp.Parsed.IsNumericRelease` is true.

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
	if p.FullName != "v0.0.1" {
		t.Fatalf("FullName = %q, want v0.0.1", p.FullName)
	}
	if p.Scope.PathPrefix != "" {
		t.Fatalf("PathPrefix = %q, want empty", p.Scope.PathPrefix)
	}
	if p.Scope.VersionPrefix != "v" {
		t.Fatalf("VersionPrefix = %q, want v", p.Scope.VersionPrefix)
	}
	if p.Version != "0.0.1" {
		t.Fatalf("Version = %q, want 0.0.1", p.Version)
	}
	if p.Prerelease != "" {
		t.Fatalf("Prerelease = %q, want empty", p.Prerelease)
	}
	if !p.IsNumericRelease {
		t.Fatal("IsNumericRelease = false, want true")
	}
}
```