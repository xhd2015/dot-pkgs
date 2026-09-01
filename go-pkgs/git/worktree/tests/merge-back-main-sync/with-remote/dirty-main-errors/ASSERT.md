# Assert

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected runner error: %v", err)
	}
	if resp.Err == "" {
		t.Fatal("expected main-sync error, got success")
	}
	if !strings.Contains(resp.Err, "main-sync") {
		t.Fatalf("error %q should mention main-sync", resp.Err)
	}
	if _, err := os.Stat(filepath.Join(req.MainRepo, "feature.txt")); err == nil {
		t.Fatal("feature.txt should not exist on main after failed sync")
	}
}
```
