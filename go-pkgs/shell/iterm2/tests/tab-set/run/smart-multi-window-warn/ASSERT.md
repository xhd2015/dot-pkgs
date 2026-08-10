## Expected

- `Warning` non-empty and mentions multiple windows (e.g. contains `2` and `window`).
- `FocusedWindow` is the first-seen window id `win-A` (most recent by find order).
- Does not create a brand-new window for the whole set (`CreatedWindow` false).

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

	if err != nil {
		t.Fatal(err)
	}
	var scripts []string
	cfg := buildRunConfig(req, &scripts)
	result, rerr := iterm2.RunTabSet(tabSetSpecFromReq(req), iterm2.TabSetRunOptions{
		Mode: iterm2.TabSetRunSmart,
	}, cfg)
	if rerr != nil {
		t.Fatalf("RunTabSet: %v", rerr)
	}
	if result == nil {
		t.Fatal("nil result")
	}
	if result.Warning == "" {
		t.Fatal("expected multi-window Warning, got empty")
	}
	w := strings.ToLower(result.Warning)
	if !strings.Contains(w, "window") {
		t.Fatalf("Warning should mention windows: %q", result.Warning)
	}
	if !strings.Contains(result.Warning, "2") && !strings.Contains(w, "multi") && !strings.Contains(w, "multiple") {
		t.Fatalf("Warning should indicate multiple windows: %q", result.Warning)
	}
	if result.FocusedWindow != "" && result.FocusedWindow != "win-A" {
		t.Fatalf("FocusedWindow = %q, want win-A (first in find order)", result.FocusedWindow)
	}
	if result.CreatedWindow {
		t.Fatal("should not CreatedWindow when sessions already found")
	}
}
```
