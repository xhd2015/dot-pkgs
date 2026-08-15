## Expected

- Reuse script scan loop reads `user.koolTargetDir` and compares it to `targetDir`.

## Exit Code

- N/A (build-script phase)

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	s := resp.Script
	if !strings.Contains(s, `variable named "user.koolTargetDir"`) {
		t.Fatalf("scan must read user.koolTargetDir: %q", s)
	}
	scanStart := strings.Index(s, "repeat with aWindow in windows")
	if scanStart < 0 {
		t.Fatal("missing window scan loop")
	}
	scanEnd := strings.Index(s[scanStart:], "if matchingWindow is not missing value then")
	if scanEnd < 0 {
		t.Fatal("missing match branch opener after scan")
	}
	scanLoop := s[scanStart : scanStart+scanEnd]
	if !strings.Contains(scanLoop, "user.koolTargetDir") {
		t.Fatalf("scan loop must check user.koolTargetDir: %q", scanLoop)
	}
	if !strings.Contains(scanLoop, "targetDir") {
		t.Fatalf("scan loop must compare against targetDir: %q", scanLoop)
	}
}
```