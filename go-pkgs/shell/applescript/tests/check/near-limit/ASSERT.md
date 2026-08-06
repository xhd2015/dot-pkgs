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
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if resp.CheckOK {
		t.Fatal("OK should be false in near_limit band")
	}
	if !resp.CheckNearLimit {
		t.Fatal("want NearLimit")
	}
	if resp.CheckSoftExceeded {
		t.Fatal("SoftExceeded should be false below SoftMax")
	}
	joined := strings.Join(resp.CheckReasons, " ")
	if !strings.Contains(joined, applescript.ReasonNearLimit) {
		t.Fatalf("reasons missing near_limit: %v", resp.CheckReasons)
	}
}
```
