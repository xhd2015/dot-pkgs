## Expected

- `StopTabSet` succeeds (nil error).
- At least one of `ClosedWindows` / `ClosedTabs` is positive.
- Captured Exec scripts contain `close` (close window or close tab).

## Exit Code

- N/A

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d

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
	if res.ClosedWindows+res.ClosedTabs < 1 {
		t.Fatalf("expected closes > 0; windows=%d tabs=%d", res.ClosedWindows, res.ClosedTabs)
	}
	all := joinedScripts(scripts)
	if !strings.Contains(strings.ToLower(all), "close") {
		t.Fatalf("Exec scripts should close sessions/windows; scripts:\n%s", all)
	}
}
```
