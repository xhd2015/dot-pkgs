## Expected

- Five entries returned.
- Entry names sorted ascending: `.codex`, `aaa-first`, `mmm-mid`, `notes.txt`, `zzz-last`.

## Errors

- `err` is nil.
- Out-of-order results or missing entries.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(resp.Entries))
	}

	var names []string
	for _, e := range resp.Entries {
		names = append(names, e.Name)
	}
	assertSortedNames(t, names)

	want := []string{".codex", "aaa-first", "mmm-mid", "notes.txt", "zzz-last"}
	for i, name := range want {
		if names[i] != name {
			t.Fatalf("entries[%d] = %q, want %q; full %v", i, names[i], name, names)
		}
	}
}
```