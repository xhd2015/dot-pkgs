## Expected

1. `err` is nil and `resp` is non-nil.
2. Snapshot `WSOutput` contains sticky footer marker `STICKY_FOOTER`.
3. Snapshot `WSOutput` contains sticky prompt marker `STICKY_PROMPT`.
4. Snapshot is a rendered frame (contains CSI hide-cursor `\x1b[?25l` or CUP),
   not an empty payload.
5. When `ExpectDirty` is set, output also contains a `DIRTY_` substring
   (latest dirty-region text may appear).

## Errors

- Missing sticky markers → live VT not used or sticky never applied.
- Empty `WSOutput` → snapshot frame not delivered.

## Side Effects

- Fixture child cleaned up (kill + session delete) via `t.Cleanup`.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	out := resp.WSOutput
	if strings.TrimSpace(out) == "" {
		t.Fatal("empty snapshot WSOutput")
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
		t.Fatalf("snapshot missing sticky marker %q (got %q)", sticky, truncate(out, 240))
	}
	if !strings.Contains(out, prompt) {
		t.Fatalf("snapshot missing prompt marker %q (got %q)", prompt, truncate(out, 240))
	}
	// Prefer rendered snapshot shape (hide cursor prefix used by renderScreenSnapshot).
	if !strings.Contains(out, "\x1b[?25l") && !strings.Contains(out, "\x1b[") {
		t.Fatalf("snapshot does not look like a CSI/CUP frame (got %q)", truncate(out, 120))
	}
	if req.ExpectDirty && !strings.Contains(out, "DIRTY_") {
		t.Fatalf("snapshot missing DIRTY_ region text (got %q)", truncate(out, 240))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
