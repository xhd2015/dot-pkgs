## Expected

- `err` is nil.
- After `ApplyLiveness`, `resp.Index.Repos` contains exactly one entry: the
  live-repo path (with `.git`).
- The dead-repo path is absent from `Repos`.
- Document-level fields (`Version`, `Universe`, `Base`) remain present on the
  returned index (filter does not wipe envelope).

## Errors

- `err` is nil.

## Side Effects

- No requirement to rewrite `repos.json` in this leaf (in-memory filter only).

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(req.Index.Repos) != 2 {
		t.Fatalf("setup seed must have 2 repos, got %d", len(req.Index.Repos))
	}

	// Seed order in SETUP: dead then live. Also verify by .git presence.
	deadPath := req.Index.Repos[0].Path
	livePath := req.Index.Repos[1].Path
	if _, err := os.Stat(filepath.Join(livePath, ".git")); err != nil {
		t.Fatalf("precondition: live path must have .git: %v", err)
	}
	if _, err := os.Stat(filepath.Join(deadPath, ".git")); err == nil {
		t.Fatalf("precondition: dead path must not have .git: %s", deadPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat dead .git: %v", err)
	}

	got := resp.Index
	if got.Version != req.Index.Version {
		t.Fatalf("Version = %d, want %d (envelope preserved)", got.Version, req.Index.Version)
	}
	if got.Universe != req.Index.Universe {
		t.Fatalf("Universe = %q, want %q", got.Universe, req.Index.Universe)
	}
	if got.Base != req.Index.Base {
		t.Fatalf("Base = %q, want %q", got.Base, req.Index.Base)
	}

	if len(got.Repos) != 1 {
		t.Fatalf("Repos len = %d, want 1 (only live), got %+v", len(got.Repos), got.Repos)
	}
	if got.Repos[0].Path != livePath {
		t.Fatalf("kept Path = %q, want live %q", got.Repos[0].Path, livePath)
	}
	for _, e := range got.Repos {
		if e.Path == deadPath {
			t.Fatalf("dead path %q must be dropped", deadPath)
		}
	}
}
```
