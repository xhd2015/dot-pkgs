## Expected

- Scan succeeds with no RootErrors; `my-repo` is discovered (warm serve).
- `resp.Stderr` contains `scan:`.
- `resp.Stderr` reports warm mode: substring `mode=warm`.
- Serve timing / counts appear: at least one of `serve`, `candidates`, `live`,
  or `duration` (contract: serve candidates/live/duration). Prefer multiple
  markers when the product logs a full serve line.
- Optional but preferred: refresh summary and/or root total markers
  (`refresh`, `total`) when present — do not fail solely on their absence if
  serve timing is clear.
- Phase-level volume only (not per-directory cold walk spam).

## Errors

- `err` is nil.

## Side Effects

- Warm path may update liveness/refresh stamps; this leaf asserts debug stderr.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if !req.Debug {
		t.Fatal("req.Debug must be true for on/warm")
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
		t.Fatalf("expected warm discovery of main repo %s, got %v", repoPath, resp.Repos)
	}

	stderr := resp.Stderr
	if !strings.Contains(stderr, "scan:") {
		t.Fatalf("Debug warm stderr must contain scan: prefix, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "mode=warm") {
		t.Fatalf("Debug warm stderr must contain mode=warm, got:\n%s", stderr)
	}

	// Serve timing / counts (contract: candidates / live / duration).
	serveMarkers := 0
	for _, m := range []string{"serve", "candidates", "live", "duration"} {
		if strings.Contains(stderr, m) {
			serveMarkers++
		}
	}
	if serveMarkers < 1 {
		t.Fatalf("Debug warm stderr must include serve timing markers (serve/candidates/live/duration), got:\n%s", stderr)
	}

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
