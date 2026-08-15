## Expected

- `notes.txt` entry: `Kind == EntryKindFile`, `Lines == "2"`.
- `binary.dat` entry: `Kind == EntryKindFile`, `Lines == "(binary)"`.
- Both entries have positive `Bytes`.

## Errors

- `err` is nil.
- Wrong line counts or missing binary marker.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/file/analyse"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.Entries))
	}

	notes := findEntry(t, resp.Entries, "notes.txt")
	if notes.Kind != analyse.EntryKindFile {
		t.Fatalf("notes.txt Kind = %q, want file", notes.Kind)
	}
	if notes.Lines != "2" {
		t.Fatalf("notes.txt Lines = %q, want 2", notes.Lines)
	}
	if notes.Bytes <= 0 {
		t.Fatalf("notes.txt Bytes = %d, want > 0", notes.Bytes)
	}

	binary := findEntry(t, resp.Entries, "binary.dat")
	if binary.Kind != analyse.EntryKindFile {
		t.Fatalf("binary.dat Kind = %q, want file", binary.Kind)
	}
	if binary.Lines != "(binary)" {
		t.Fatalf("binary.dat Lines = %q, want (binary)", binary.Lines)
	}
	if binary.Bytes <= 0 {
		t.Fatalf("binary.dat Bytes = %d, want > 0", binary.Bytes)
	}
}
```