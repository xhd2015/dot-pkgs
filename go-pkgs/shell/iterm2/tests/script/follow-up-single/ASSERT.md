## Expected

- Script contains cd line and `write text "grok"`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Script, `write text ("cd " & quoted form of targetDir)`) {
		t.Fatal("missing cd")
	}
	if !strings.Contains(resp.Script, `write text "grok"`) {
		t.Fatalf("missing grok follow-up: %q", resp.Script)
	}
}
```