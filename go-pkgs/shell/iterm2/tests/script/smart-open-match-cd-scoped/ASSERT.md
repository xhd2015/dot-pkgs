## Expected

- Match branch scopes `write text` cd to the tab created in `matchingWindow`.

## Exit Code

- N/A (build-script phase)

```go
import (
	"strings"
	"testing"
)

func smartScriptMatchBranch(script string) string {
	const open = `if matchingWindow is not missing value then`
	start := strings.Index(script, open)
	if start < 0 {
		return ""
	}
	rest := script[start+len(open):]
	elseIdx := strings.Index(rest, "\n  else\n")
	if elseIdx < 0 {
		elseIdx = strings.Index(rest, "\n  else")
	}
	if elseIdx < 0 {
		return rest
	}
	return rest[:elseIdx]
}

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	s := resp.Script
	match := smartScriptMatchBranch(s)
	if match == "" {
		t.Fatalf("missing match branch: %q", s)
	}
	if strings.Contains(match, "tell current session of current tab\n") ||
		strings.Contains(match, "tell current session of current tab\r\n") {
		t.Fatalf("must not cd via unqualified current tab: %q", match)
	}
	scoped := strings.Contains(match, "current session of current tab of matchingWindow") ||
		strings.Contains(match, "current session of newTab")
	if !scoped {
		t.Fatalf("must scope cd to matchingWindow tab: %q", match)
	}
}
```