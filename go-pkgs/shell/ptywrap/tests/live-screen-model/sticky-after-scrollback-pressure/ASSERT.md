## Expected

1. `err` is nil and `resp` is non-nil.
2. After ring pressure, snapshot `WSOutput` **still** contains `STICKY_FOOTER`.
3. Snapshot also contains `STICKY_PROMPT` (bottom chrome pair).
4. Snapshot is non-empty and uses CSI/CUP framing (`\x1b[?25l` preferred).

## Errors

- Missing `STICKY_FOOTER` after pressure is the production failure class:
  cold `renderScreenSnapshot(scrollback)` after ring trim drops early sticky
  paint. Fix: export cells from the persistent live VT instead.

## Side Effects

- Emits ≥ 320 KiB of PTY output (CI cost ~1–5s typical); no special label
  unless measured slow on target runners.
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
		t.Fatal("empty snapshot WSOutput after scrollback pressure")
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
		t.Fatalf("RED-defining: snapshot missing sticky %q after ring pressure (cold replay of trimmed scrollback loses chrome); got %q",
			sticky, truncate(out, 280))
	}
	if !strings.Contains(out, prompt) {
		t.Fatalf("snapshot missing prompt marker %q after ring pressure (got %q)",
			prompt, truncate(out, 280))
	}
	if !strings.Contains(out, "\x1b[?25l") && !strings.Contains(out, "\x1b[") {
		t.Fatalf("snapshot does not look like a CSI/CUP frame (got %q)", truncate(out, 120))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
```
