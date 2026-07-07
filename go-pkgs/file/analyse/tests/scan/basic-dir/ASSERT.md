## Expected

- Exactly two entries: `notes.txt` (file) and `plain-dir` (dir).
- `plain-dir` has `Kind == EntryKindDir`.
- `plain-dir` children sorted alphabetically; includes `sub` with non-zero bytes.
- `plain-dir.Bytes` equals sum of deep child sizes (at least `sub` content).

## Errors

- `err` is nil.
- Missing or wrong entry kind.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/analyse"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.Entries))
	}

	plain := findEntry(t, resp.Entries, "plain-dir")
	if plain.Kind != analyse.EntryKindDir {
		t.Fatalf("plain-dir Kind = %q, want dir", plain.Kind)
	}
	if plain.Bytes <= 0 {
		t.Fatalf("plain-dir Bytes = %d, want > 0", plain.Bytes)
	}

	names := childNames(plain)
	assertSortedNames(t, names)
	if len(names) != 1 || names[0] != "sub" {
		t.Fatalf("plain-dir children = %v, want [sub]", names)
	}
	if plain.Children[0].Bytes <= 0 {
		t.Fatalf("sub child bytes = %d, want > 0", plain.Children[0].Bytes)
	}

	notes := findEntry(t, resp.Entries, "notes.txt")
	if notes.Kind != analyse.EntryKindFile {
		t.Fatalf("notes.txt Kind = %q, want file", notes.Kind)
	}
}
```