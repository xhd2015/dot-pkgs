## Expected

- `StopTabSet` returns nil error.
- `ClosedWindows == 0` and `ClosedTabs == 0`.
- `Warning` non-empty (not running / no sessions).

## Exit Code

- N/A

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	var scripts []string
	res, serr := iterm2.StopTabSet(req.TabSetName, buildStopConfig(req, &scripts))
	if serr != nil {
		t.Fatalf("StopTabSet: %v", serr)
	}
	if res == nil {
		t.Fatal("nil stop result")
	}
	if res.ClosedWindows != 0 || res.ClosedTabs != 0 {
		t.Fatalf("closed windows=%d tabs=%d, want 0/0", res.ClosedWindows, res.ClosedTabs)
	}
	if res.Warning == "" {
		t.Fatal("expected not-running Warning when find empty")
	}
}
```
