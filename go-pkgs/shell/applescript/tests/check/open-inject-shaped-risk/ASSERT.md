## Expected

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp.CheckByteLen <= applescript.WriteTextSoftMaxBytes {
		t.Fatalf("setup should exceed SoftMax: byteLen=%d", resp.CheckByteLen)
	}
	if !resp.CheckSoftExceeded || resp.CheckOK {
		t.Fatalf("want SoftExceeded and !OK soft=%v ok=%v", resp.CheckSoftExceeded, resp.CheckOK)
	}
	// Escape still defined for same input (no panic path)
	_ = applescript.EscapeString(req.CheckInput)
	if !strings.Contains(req.CheckInput, "__seq_1") || !strings.Contains(req.CheckInput, "字") {
		t.Fatal("fixture should include __seq_ and Chinese")
	}
}
```
