## Expected

- Scan succeeds (warm or cold fall-back ok for this leaf).
- Path `<CacheRoot>/mirror` does **not** exist after the Scan under test.
- Known soft-omit: brand-new under non-sibling unit need not appear in Result.

## Errors

- `err` is nil.

## Side Effects

- Warm path must not invent `mirror/` for liveness or unit refresh.
- **RED** while product still writes mirror on cold seed and/or warm rewalk.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	mirrorDir := filepath.Join(req.CacheRoot, "mirror")
	if st, statErr := os.Stat(mirrorDir); statErr == nil {
		t.Fatalf("warm path left mirror at %s (mode=%v); mirror cache is retired — want path absent", mirrorDir, st.Mode())
	} else if !os.IsNotExist(statErr) {
		t.Fatalf("stat mirror: %v", statErr)
	}
}
```
