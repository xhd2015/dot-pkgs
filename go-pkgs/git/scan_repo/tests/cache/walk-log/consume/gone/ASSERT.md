## Expected

- Cold visited `notes` (path present among cold-era visit events before
  second Scan appends — after Run, visit for notes still appears earlier in
  the log).
- After second Scan, at least one event with `op=gone` whose `path` is the
  abs path of the deleted `notes` directory.
- The gone event appears **after** the first `gen_end` gen=1 (appended by
  consume, not during cold).
- Prefer also gen_end 2 present (full consume cycle); if budget only partial,
  gone alone is the primary contract for this leaf.

## Errors

- `err` is nil.

## Side Effects

- Append-only log gains a gone line for the removed directory.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if !resp.WalkLogOK {
		t.Fatal("expected walk.jsonl after consume")
	}

	notes := absPath(t, filepath.Join(req.Roots[0], "notes"))

	// Locate first gen_end gen=1 index.
	gen1Idx := -1
	for i, ev := range resp.WalkEvents {
		if ev.Op == "gen_end" && ev.Gen == 1 {
			gen1Idx = i
			break
		}
	}
	if gen1Idx < 0 {
		t.Fatalf("missing gen_end gen=1; events=%v", resp.WalkEvents)
	}

	// Cold portion should have visited notes (before gen_end 1).
	visitedNotes := false
	for i := 0; i < gen1Idx; i++ {
		ev := resp.WalkEvents[i]
		if ev.Op == "visit" && absPath(t, ev.Path) == notes {
			visitedNotes = true
			break
		}
	}
	if !visitedNotes {
		t.Fatalf("cold portion missing visit for notes %q; events=%v", notes, resp.WalkEvents)
	}

	// Consume appends gone after gen_end 1.
	foundGone := false
	for i := gen1Idx + 1; i < len(resp.WalkEvents); i++ {
		ev := resp.WalkEvents[i]
		if ev.Op == "gone" && ev.Path != "" && absPath(t, ev.Path) == notes {
			foundGone = true
			break
		}
	}
	if !foundGone {
		t.Fatalf("expected gone event for %q after gen_end 1; events=%v",
			notes, resp.WalkEvents)
	}
}
```
