## Expected

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(resp.Doc) == "" {
		t.Fatal("empty DocumentWriteTextLimitation")
	}
	lower := strings.ToLower(resp.Doc)
	for _, kw := range []string{"write text", "1024", "900", "prompt", "file"} {
		if !strings.Contains(lower, strings.ToLower(kw)) && kw != "prompt" {
			// prompt-file may be phrased as "on disk" / "script"
			continue
		}
	}
	if !strings.Contains(lower, "write text") {
		t.Fatal("doc must mention write text")
	}
	if !strings.Contains(resp.Doc, "1024") && !strings.Contains(resp.Doc, "SoftMax") {
		t.Fatal("doc must mention soft max / 1024")
	}
	if !strings.Contains(resp.Doc, "900") && !strings.Contains(resp.Doc, "SafeMax") {
		t.Fatal("doc must mention safe max / 900")
	}
	if !strings.Contains(lower, "disk") && !strings.Contains(lower, "file") && !strings.Contains(lower, "script") {
		t.Fatal("doc must mention file/script workaround")
	}
}
```
