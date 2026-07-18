## Expected

- After cold: `ColdCursorOffset == ColdWalkLogSize` and both &gt; 0.
- After second Scan: `CursorOK`, `Cursor.Offset == WalkLogSize`, and
  `Cursor.Offset > ColdCursorOffset` (cursor advanced past the cold seal).
- Log still ends with gen_end 2 (consume completed the generation cycle).

## Errors

- `err` is nil.

## Side Effects

- Durable `home/walk.cursor.json` rewritten to the post-consume EOF.

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if !resp.ColdCursorOK || resp.ColdCursorOffset <= 0 {
		t.Fatalf("cold cursor not sealed: OK=%v offset=%d",
			resp.ColdCursorOK, resp.ColdCursorOffset)
	}
	if resp.ColdCursorOffset != resp.ColdWalkLogSize {
		t.Fatalf("cold Cursor.Offset=%d != ColdWalkLogSize=%d",
			resp.ColdCursorOffset, resp.ColdWalkLogSize)
	}

	if !resp.CursorOK {
		t.Fatal("expected CursorOK after consume")
	}
	wantCur := homeWalkCursorPath(req.CacheRoot)
	if _, statErr := os.Stat(wantCur); statErr != nil {
		t.Fatalf("walk.cursor.json missing after consume: %v", statErr)
	}
	if resp.Cursor.Offset <= 0 {
		t.Fatalf("Cursor.Offset = %d, want > 0", resp.Cursor.Offset)
	}
	if resp.Cursor.Offset != resp.WalkLogSize {
		t.Fatalf("Cursor.Offset = %d, want WalkLogSize %d (post-consume EOF)",
			resp.Cursor.Offset, resp.WalkLogSize)
	}
	if resp.Cursor.Offset <= resp.ColdCursorOffset {
		t.Fatalf("Cursor.Offset = %d, want > ColdCursorOffset %d (must advance)",
			resp.Cursor.Offset, resp.ColdCursorOffset)
	}

	last, ok := lastGenEnd(resp.WalkEvents)
	if !ok || last.Gen != 2 {
		t.Fatalf("expected last gen_end gen=2 for cursor advance leaf; last=%v ok=%v events=%v",
			last, ok, resp.WalkEvents)
	}
}
```
