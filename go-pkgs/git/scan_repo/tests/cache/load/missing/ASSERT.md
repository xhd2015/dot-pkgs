## Expected

- `resp.EntryOK` is false.
- `resp.Entry` is the zero `CacheEntry` (version 0, empty strings, false flags).

## Errors

- `err` is nil (missing is not an error).

```go
import (
	"reflect"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("expected nil error for missing entry, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.EntryOK {
		t.Fatal("expected EntryOK false when entry.json is missing")
	}
	zero := scan_repo.CacheEntry{}
	if !reflect.DeepEqual(resp.Entry, zero) {
		t.Fatalf("expected zero CacheEntry, got %+v", resp.Entry)
	}
}
```

