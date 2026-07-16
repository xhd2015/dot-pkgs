## Expected

- Scan succeeds (`err` nil) with no RootErrors.
- At least one main repo discovered (`my-repo`).
- `resp.Stderr` contains the greppable prefix `scan:`.
- `resp.Stderr` reports cold mode: substring `mode=cold` (or an equivalent
  `mode=` token whose value is `cold`).
- Prefer a cold reason token such as `missing_root_entry` (empty root cache).
  Accept `scan_complete_false` if the product uses that wording for the same
  empty-cache case. Substring match is enough; exact line layout is product-owned.
- Stderr must not look like per-directory walk spam: line count of lines
  containing `scan:` stays small (phase-level), not one line per walked dir.

## Errors

- `err` is nil.

## Side Effects

- Mirror may be written (cold write); this leaf asserts debug stderr only.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if !req.Debug {
		t.Fatal("req.Debug must be true for on/cold")
	}
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}

	repoPath := absPath(t, filepath.Join(req.Roots[0], "my-repo"))
	var found bool
	for _, r := range resp.Repos {
		if r.Path == repoPath && r.RepoType == scan_repo.RepoTypeMain {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected discovery of main repo %s, got %v", repoPath, resp.Repos)
	}

	stderr := resp.Stderr
	if !strings.Contains(stderr, "scan:") {
		t.Fatalf("Debug cold stderr must contain scan: prefix, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "mode=cold") {
		t.Fatalf("Debug cold stderr must contain mode=cold, got:\n%s", stderr)
	}
	// Reason for empty / missing root cache (contract examples).
	hasReason := strings.Contains(stderr, "missing_root_entry") ||
		strings.Contains(stderr, "scan_complete_false") ||
		strings.Contains(stderr, "reason=")
	if !hasReason {
		t.Fatalf("Debug cold stderr should include a cold reason (e.g. missing_root_entry), got:\n%s", stderr)
	}

	// Phase-level volume: not one scan: line per directory in the tree.
	scanLines := 0
	for _, line := range strings.Split(stderr, "\n") {
		if strings.Contains(line, "scan:") {
			scanLines++
		}
	}
	if scanLines > 40 {
		t.Fatalf("too many scan: lines (%d); want phase-level logs not per-dir spam:\n%s", scanLines, stderr)
	}
}
```
