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
	if !resp.CheckOK {
		t.Fatalf("CheckOK=false reasons=%v byteLen=%d", resp.CheckReasons, resp.CheckByteLen)
	}
	if resp.CheckSoftExceeded || resp.CheckNearLimit {
		t.Fatalf("unexpected soft/near flags soft=%v near=%v", resp.CheckSoftExceeded, resp.CheckNearLimit)
	}
	if resp.CheckByteLen > applescript.WriteTextSafeMaxBytes {
		t.Fatalf("byteLen %d > SafeMax %d", resp.CheckByteLen, applescript.WriteTextSafeMaxBytes)
	}
}
```

