## Expected

1. `err` is nil and `resp` is non-nil.
2. Snapshot `WSOutput` contains `STICKY_FOOTER` after resize + sticky paint.
3. Snapshot contains `STICKY_PROMPT`.
4. Snapshot non-empty with CSI/CUP framing.
5. When `ExpectDirty` is set, a `DIRTY_` substring is present.

## Errors

- Sticky missing after resize → live VT not resized/recreated with geometry, or
  fixture painted before resize landed.
- Empty frame → snapshot export failure post-resize.

## Side Effects

- Writer attach used only for resize; final observations use `attach_mode=snapshot`.
- Fixture child cleaned up via `t.Cleanup`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	out := resp.WSOutput
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty snapshot WSOutput after resize+sticky")
	}
	sticky := req.StickyMarker
	if sticky == "" {
		sticky = "STICKY_FOOTER"
	}
	prompt := req.PromptMarker
	if prompt == "" {
		prompt = "STICKY_PROMPT"
	}
	if !strings.Contains(out, sticky) {
		t.Fatalf("snapshot missing sticky %q after resize (got %q)", sticky, truncate(out, 240))
	}
	if !strings.Contains(out, prompt) {
		t.Fatalf("snapshot missing prompt %q after resize (got %q)", prompt, truncate(out, 240))
	}
	if !strings.Contains(out, "\x1b[?25l") && !strings.Contains(out, "\x1b[") {
		t.Fatalf("snapshot does not look like a CSI/CUP frame (got %q)", truncate(out, 120))
	}
	if req.ExpectDirty && !strings.Contains(out, "DIRTY_") {
		t.Fatalf("snapshot missing DIRTY_ text after resize (got %q)", truncate(out, 240))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
