## Expected

- Script must not contain `exec $SHELL`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(resp.Script, "exec $SHELL") {
		t.Fatalf("script must not exec shell: %q", resp.Script)
	}
}
```