## Expected

- Script scans session `path` and tracks matching tab/session.
- On match, focuses via `select` without `write text` cd in the match branch.
- Miss branch creates window and cds; must not use `current session of current tab of current window`.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	s := resp.Script
	if !strings.Contains(s, `variable named "path"`) {
		t.Fatalf("reuse script must scan session path: %q", s)
	}
	if !strings.Contains(s, "matchingTab") && !strings.Contains(s, "matchingSession") {
		t.Fatal("reuse script must track matching tab/session")
	}
	if strings.Contains(s, "current session of current tab of current window") {
		t.Fatal("reuse script must not cd in arbitrary current session")
	}
	if !strings.Contains(s, "create window with default profile") {
		t.Fatal("reuse script must support new-window miss branch")
	}
	if !strings.Contains(s, `write text ("cd " & quoted form of targetDir)`) {
		t.Fatalf("miss branch must cd via quoted form of targetDir: %q", s)
	}
}
```