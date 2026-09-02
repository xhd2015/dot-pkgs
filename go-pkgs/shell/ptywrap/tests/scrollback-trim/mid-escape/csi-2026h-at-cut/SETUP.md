# Scenario

**Bug**: raw `buf[len-max:]` cut on the `0` of `\x1b[?2026h` leaves leading `026h`

```
# crime scene (Codex synchronized-update CSI)
prefix + ESC[?2026h + TAIL_MARKER + suffix
  provisional cut = index of '0' in "026h"
  -> trimScrollback(Data, Max)
  -> must NOT start with "026h"
  -> must still contain "TAIL_MARKER"
```

## Steps

1. Build buffer: `AAAA…` + `\x1b[?2026h` + `TAIL_MARKER\n` + `BBBB…`.
2. Choose `Max` so provisional cut lands on byte index `len(prefix)+4`
   (`\x1b[?2` is 4 bytes; next is `0` of `026h`).
3. Call trim via root `Run`.

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	prefix := bytes.Repeat([]byte("A"), 32)
	seq := []byte("\x1b[?2026h") // indices 0..3 = ESC[?2 ; 4..7 = 026h
	tail := []byte("TAIL_MARKER\n")
	suffix := bytes.Repeat([]byte("B"), 48)
	req.Data = append(append(append(prefix, seq...), tail...), suffix...)
	// cut at prefix+4 → kept would start with "026h" under raw slice
	cut := len(prefix) + 4
	req.Max = len(req.Data) - cut
	if req.Max <= 0 {
		t.Fatalf("bad Max=%d len=%d cut=%d", req.Max, len(req.Data), cut)
	}
	return nil
}
```
