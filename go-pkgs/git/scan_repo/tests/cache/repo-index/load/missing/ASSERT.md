## Expected

- `resp.IndexOK` is false.
- `resp.Index` is the zero / empty `RepoIndex` (version 0, empty universe/base,
  nil or empty `Repos`).
- Missing file is not treated as corrupt.

## Errors

- `err` is nil (missing is not an error).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected nil error for missing index, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.IndexOK {
		t.Fatal("expected IndexOK false when repos.json is missing")
	}
	got := resp.Index
	if got.Version != 0 {
		t.Fatalf("Version = %d, want 0 for missing", got.Version)
	}
	if got.Universe != "" {
		t.Fatalf("Universe = %q, want empty for missing", got.Universe)
	}
	if got.Base != "" {
		t.Fatalf("Base = %q, want empty for missing", got.Base)
	}
	if got.UpdatedAt != "" {
		t.Fatalf("UpdatedAt = %q, want empty for missing", got.UpdatedAt)
	}
	if len(got.Repos) != 0 {
		t.Fatalf("Repos = %+v, want empty for missing", got.Repos)
	}
	// Zero-value structural check (Repos nil or empty both ok).
	_ = scan_repo.RepoIndex{}
}
```
