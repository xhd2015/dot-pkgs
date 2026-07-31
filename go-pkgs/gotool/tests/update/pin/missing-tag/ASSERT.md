## Expected

- `Pin` returns an error (missing latest tag when Version empty).
- Error text mentions tag (e.g. "no tag" / "tag") so a generic "not implemented" stub stays RED.
- Consumer go.mod unchanged: require still `v0.0.1`, replace still present.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err == nil {
		t.Fatal("Pin with untagged DepDir and empty Version: expected error, got nil")
	}
	errText := strings.ToLower(resp.Err.Error())
	if !strings.Contains(errText, "tag") {
		t.Fatalf("error %q should mention tag (missing version tag)", resp.Err)
	}
	if resp.DiskVersion != "v0.0.1" {
		t.Fatalf("disk require version = %q, want v0.0.1 (unchanged on error)", resp.DiskVersion)
	}
	if !resp.HasReplace {
		t.Fatalf("replace for %q must remain when Pin errors", fixtureModulePath)
	}
}
```
