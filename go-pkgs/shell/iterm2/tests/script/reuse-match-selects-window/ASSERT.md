## Expected

- Reuse match branch calls `select matchingWindow` at application level.

## Exit Code

- N/A (build-script phase)

```go
import (
	"strings"
	"testing"
)

func reuseScriptMatchBranch(script string) string {
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
	if !strings.Contains(s, "select matchingWindow") {
		t.Fatalf("reuse match must select matchingWindow at app level: %q", s)
	}
	match := reuseScriptMatchBranch(s)
	if match == "" {
		t.Fatalf("missing match branch: %q", s)
	}
	if strings.Contains(match, "write text") {
		t.Fatalf("match branch must not cd: %q", match)
	}
}
```