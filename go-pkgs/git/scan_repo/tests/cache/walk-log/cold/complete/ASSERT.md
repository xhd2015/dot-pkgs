## Expected

- Scan succeeds with no RootErrors; Result includes main `projects/alpha`.
- `resp.WalkLogOK` is true; file exists at `<CacheRoot>/home/walk.jsonl`.
- Parsed events include at least one `op=visit` whose `path` is the abs scan
  root and visits covering intermediate dirs (`notes`, `projects`) and the
  repo dir (`projects/alpha`) — assert by path membership of visit events.
- The **last** event is `op=gen_end` with `gen=1` (successful full cold seal;
  gen_end present because cold completed — incomplete cold not tested here).
- `resp.CursorOK` is true; file exists at `<CacheRoot>/home/walk.cursor.json`.
- `resp.Cursor.Offset` equals `resp.WalkLogSize` (byte offset at sealed EOF)
  and is strictly greater than zero.

## Errors

- `err` is nil.

## Side Effects

- Durable walk log + cursor under universe `home` after first cold Scan.
- Index / mirror side effects from P1–P2 may also exist; this leaf only
  asserts walk log contract.

```go
import (
	"os"
	"path/filepath"
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
	if len(resp.RootErrors) != 0 {
		t.Fatalf("expected no RootErrors, got %v", resp.RootErrors)
	}

	rootAbs := absPath(t, req.Roots[0])
	notes := absPath(t, filepath.Join(rootAbs, "notes"))
	projects := absPath(t, filepath.Join(rootAbs, "projects"))
	alpha := absPath(t, filepath.Join(rootAbs, "projects", "alpha"))

	// Discovery still works (P0/P2 unchanged).
	foundAlpha := false
	for _, r := range resp.Repos {
		if absPath(t, r.Path) == alpha && r.RepoType == scan_repo.RepoTypeMain {
			foundAlpha = true
		}
	}
	if !foundAlpha {
		t.Fatalf("Result missing main alpha %q; repos=%v", alpha, resp.Repos)
	}

	wantLog := homeWalkLogPath(req.CacheRoot)
	wantCur := homeWalkCursorPath(req.CacheRoot)
	if resp.WalkLogPath != wantLog {
		t.Fatalf("WalkLogPath = %q, want %q", resp.WalkLogPath, wantLog)
	}
	if resp.CursorPath != wantCur {
		t.Fatalf("CursorPath = %q, want %q", resp.CursorPath, wantCur)
	}

	if !resp.WalkLogOK {
		t.Fatal("expected WalkLogOK=true after cold Scan writes home/walk.jsonl")
	}
	if _, statErr := os.Stat(wantLog); statErr != nil {
		t.Fatalf("home/walk.jsonl missing after cold Scan: %v", statErr)
	}
	if resp.WalkLogSize <= 0 {
		t.Fatalf("WalkLogSize = %d, want > 0", resp.WalkLogSize)
	}
	if len(resp.WalkEvents) == 0 {
		t.Fatal("expected non-empty walk.jsonl events")
	}

	// Collect visit paths (normalize abs).
	visitPaths := make(map[string]struct{})
	for i, ev := range resp.WalkEvents {
		if ev.Op == "visit" {
			if ev.Path == "" {
				t.Fatalf("WalkEvents[%d]: visit missing path", i)
			}
			visitPaths[absPath(t, ev.Path)] = struct{}{}
		}
	}
	if len(visitPaths) == 0 {
		t.Fatalf("expected at least one visit event; events=%v", resp.WalkEvents)
	}

	// Must visit scan root and fixture dirs (dirs walked during cold).
	for _, p := range []string{rootAbs, notes, projects, alpha} {
		if _, ok := visitPaths[p]; !ok {
			t.Fatalf("visit events missing path %q; visits=%v events=%v", p, visitPaths, resp.WalkEvents)
		}
	}

	// Seal: last event is gen_end with gen=1 (full cold success).
	last := resp.WalkEvents[len(resp.WalkEvents)-1]
	if last.Op != "gen_end" {
		t.Fatalf("last event op = %q, want gen_end; events=%v", last.Op, resp.WalkEvents)
	}
	if last.Gen != 1 {
		t.Fatalf("gen_end.gen = %d, want 1", last.Gen)
	}
	// Exactly one gen_end for first cold (pragmatic: at least last is gen_end 1).
	genEndCount := 0
	for _, ev := range resp.WalkEvents {
		if ev.Op == "gen_end" {
			genEndCount++
			if ev.Gen != 1 {
				t.Fatalf("gen_end.gen = %d, want 1 for first cold", ev.Gen)
			}
		}
	}
	if genEndCount != 1 {
		t.Fatalf("gen_end count = %d, want 1 after single cold Scan", genEndCount)
	}

	if !resp.CursorOK {
		t.Fatal("expected CursorOK=true after cold Scan writes home/walk.cursor.json")
	}
	if _, statErr := os.Stat(wantCur); statErr != nil {
		t.Fatalf("home/walk.cursor.json missing after cold Scan: %v", statErr)
	}
	if resp.Cursor.Offset <= 0 {
		t.Fatalf("Cursor.Offset = %d, want > 0", resp.Cursor.Offset)
	}
	if resp.Cursor.Offset != resp.WalkLogSize {
		t.Fatalf("Cursor.Offset = %d, want WalkLogSize %d (sealed EOF)",
			resp.Cursor.Offset, resp.WalkLogSize)
	}
}
```
