# Scenario

**Feature**: valid UTF-8 with a multi-byte rune straddling the 512-byte sniff window is text

```
│ + pad + U+2591 split at offset 512 + tail -> DetectFileType -> isBinary=false
```

## Steps

1. Build a synthetic file that is fully valid UTF-8 overall, starts with box-drawing
   `│` (`E2 94 82`), and places the 3-byte light-shade rune `░` (`E2 96 91`) so the
   first 512 sniff bytes end mid-sequence (bytes 510–511 = `E2 96`, continuation in
   the rest of the file).
2. Write to a temp path and set `req.Path`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Sniff window is 512 bytes. Build content so a 3-byte UTF-8 sequence starts
	// at offset 510 and is completed after the window (valid UTF-8 file overall).
	//
	// Layout of first 512 bytes:
	//   [0:3)   │  = E2 94 82
	//   [3:510) ASCII 'a' padding (507 bytes)
	//   [510:512) first two bytes of ░ (E2 96) — incomplete only at sniff boundary
	// Rest of file: 0x91 (completes ░) + more ASCII.
	const sniff = 512
	prefix := []byte{0xE2, 0x94, 0x82} // │
	padLen := sniff - len(prefix) - 2  // leave 2 bytes for incomplete lead of ░
	if padLen < 0 {
		t.Fatal("fixture layout math error")
	}
	data := make([]byte, 0, sniff+16)
	data = append(data, prefix...)
	for i := 0; i < padLen; i++ {
		data = append(data, 'a')
	}
	// First two bytes of U+2591 LIGHT SHADE (E2 96 91)
	data = append(data, 0xE2, 0x96)
	// Complete the rune + trailing text so the whole file is valid UTF-8
	data = append(data, 0x91)
	data = append(data, []byte(" tail\n")...)

	if len(data) <= sniff {
		t.Fatalf("fixture must exceed sniff window, got len=%d", len(data))
	}
	req.Path = writeTempFile(t, "utf8-trunc.bin.txt", data)
	return nil
}
```
