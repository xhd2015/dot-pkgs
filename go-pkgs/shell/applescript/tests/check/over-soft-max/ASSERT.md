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
	if resp.CheckOK {
		t.Fatal("OK should be false over SoftMax")
	}
	if !resp.CheckSoftExceeded {
		t.Fatal("want SoftExceeded")
	}
	joined := strings.Join(resp.CheckReasons, " ")
	if !strings.Contains(joined, applescript.ReasonSoftMaxBytes) {
		t.Fatalf("reasons missing soft_max_bytes: %v", resp.CheckReasons)
	}
}
```
