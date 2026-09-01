# Assert

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Err != "" {
		t.Fatalf("unexpected error: %s", resp.Err)
	}
	if resp.Action != "merged" {
		t.Fatalf("Action=%q want merged", resp.Action)
	}
	featurePath := filepath.Join(req.MainRepo, "feature.txt")
	if _, err := os.Stat(featurePath); err != nil {
		t.Fatalf("feature.txt missing on main after land: %v", err)
	}
}
```
