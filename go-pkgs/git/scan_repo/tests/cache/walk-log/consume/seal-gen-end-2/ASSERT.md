## Expected

- Run succeeds (`err` nil); cold left a walk log (`ColdWalkLogOK`,
  `ColdCursorOK`) and the final log remains readable (`WalkLogOK`).
- Events include exactly one `gen_end` with `gen=1` and at least one
  `gen_end` with `gen=2`.
- The **last** `gen_end` in the log has `gen=2` (second Scan sealed G+1 after
  consuming gen_end 1).
- Discovery of main `projects/alpha` still works on the second Scan result.

## Errors

- `err` is nil.

## Side Effects

- `home/walk.jsonl` grows (or at least gains gen_end 2) after the second Scan.
- Cursor file remains present (detailed offset checks live in `cursor-advance`).

```go
import (
	"path/filepath"
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
	if !resp.ColdWalkLogOK || !resp.ColdCursorOK {
		t.Fatalf("cold seal missing: WalkLogOK=%v CursorOK=%v (P3 prerequisite)",
			resp.ColdWalkLogOK, resp.ColdCursorOK)
	}
	if !resp.WalkLogOK {
		t.Fatal("expected walk.jsonl after consume Scan")
	}
	if len(resp.WalkEvents) == 0 {
		t.Fatal("expected walk events after consume")
	}

	if countGenEnd(resp.WalkEvents, 1) != 1 {
		t.Fatalf("gen_end gen=1 count = %d, want 1; events=%v",
			countGenEnd(resp.WalkEvents, 1), resp.WalkEvents)
	}
	if countGenEnd(resp.WalkEvents, 2) < 1 {
		t.Fatalf("expected at least one gen_end gen=2 after second Scan; events=%v",
			resp.WalkEvents)
	}
	last, ok := lastGenEnd(resp.WalkEvents)
	if !ok {
		t.Fatal("no gen_end events in walk.jsonl")
	}
	if last.Gen != 2 {
		t.Fatalf("last gen_end.gen = %d, want 2 (consume seals G+1); events=%v",
			last.Gen, resp.WalkEvents)
	}

	alpha := absPath(t, filepath.Join(req.Roots[0], "projects", "alpha"))
	found := false
	for _, r := range resp.Repos {
		if absPath(t, r.Path) == alpha && r.RepoType == scan_repo.RepoTypeMain {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("second Scan missing main alpha %q; repos=%v", alpha, resp.Repos)
	}
}
```
