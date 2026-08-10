## Expected

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if resp.CheckByteLen != applescript.WriteTextSafeMaxBytes {
		t.Fatalf("byteLen=%d want %d", resp.CheckByteLen, applescript.WriteTextSafeMaxBytes)
	}
	if !resp.CheckOK {
		t.Fatalf("want OK at SafeMax reasons=%v", resp.CheckReasons)
	}
}
```
