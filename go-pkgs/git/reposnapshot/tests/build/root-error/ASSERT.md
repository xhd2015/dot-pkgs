## Expected

- `Snapshot.RootErrors` has one entry with path `missing-scan-root`.
- `Nodes` has one synthetic node for `missing-scan-root`.
- Synthetic node `Error` is `scan failed: stat: no such file or directory`.
- Checkout fields on synthetic node are empty.

## Errors

- `err` is nil.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Snapshot.RootErrors) != 1 {
		t.Fatalf("expected 1 RootError, got %+v", resp.Snapshot.RootErrors)
	}
	re := resp.Snapshot.RootErrors[0]
	if re.Path != "missing-scan-root" {
		t.Fatalf("RootErrors[0].Path = %q, want missing-scan-root", re.Path)
	}
	if !strings.Contains(re.Error, "no such file") {
		t.Fatalf("RootErrors[0].Error = %q, want stat failure", re.Error)
	}

	if len(resp.Snapshot.Nodes) != 1 {
		t.Fatalf("expected 1 synthetic node, got %d", len(resp.Snapshot.Nodes))
	}
	node := resp.Snapshot.Nodes[0]
	if node.Path != "missing-scan-root" {
		t.Fatalf("node.Path = %q, want missing-scan-root", node.Path)
	}
	wantErr := "scan failed: stat: no such file or directory"
	if node.Error != wantErr {
		t.Fatalf("node.Error = %q, want %q", node.Error, wantErr)
	}
	if node.Checkout.Branch != "" || node.Checkout.CommitSHA != "" ||
		node.Checkout.CommitMsg != "" || node.Checkout.Status != "" || node.Checkout.Error != "" {
		t.Fatalf("expected empty checkout on synthetic node, got %+v", node.Checkout)
	}
}
```
