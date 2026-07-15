## Expected

- `resp.EntryOK` is true.
- Loaded entry matches `req.EntryB` (second save), not `req.Entry`.
- On-disk `entry.json` is valid JSON that unmarshals to the same second entry.

## Errors

- `err` is nil.

## Side Effects

- Final `entry.json` at the expected mirror path is parseable JSON (no truncated
  half-write left as the only content).

```go
import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if !resp.EntryOK {
		t.Fatal("expected EntryOK true after overwrite")
	}
	if !reflect.DeepEqual(resp.Entry, req.EntryB) {
		t.Fatalf("loaded entry = %+v, want EntryB %+v", resp.Entry, req.EntryB)
	}
	if reflect.DeepEqual(resp.Entry, req.Entry) {
		t.Fatal("loaded entry still matches first write; expected second writer to win")
	}

	path := expectedMirrorEntryPath(t, req.CacheRoot, req.RealPath)
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("read entry.json: %v", readErr)
	}
	var disk scan_repo.CacheEntry
	if jsonErr := json.Unmarshal(raw, &disk); jsonErr != nil {
		t.Fatalf("entry.json is not valid JSON: %v\n%s", jsonErr, raw)
	}
	if !reflect.DeepEqual(disk, req.EntryB) {
		t.Fatalf("on-disk entry = %+v, want EntryB %+v", disk, req.EntryB)
	}
}
```
