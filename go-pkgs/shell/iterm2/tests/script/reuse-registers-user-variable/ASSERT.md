## Expected

- Reuse script miss branch sets `user.koolTargetDir` to `targetDir` after `cd`.
- Match branch must not register the user variable.

## Exit Code

- N/A (build-script phase)

```go
import (
	"strings"
	"testing"
)

const koolTargetVar = `variable named "user.koolTargetDir"`

func reuseScriptElseBranch(script string) string {
	const marker = `create window with default profile`
	idx := strings.Index(script, marker)
	if idx < 0 {
		return ""
	}
	return script[idx:]
}

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
	elseBranch := reuseScriptElseBranch(s)
	if elseBranch == "" {
		t.Fatalf("missing else branch: %q", s)
	}
	if !strings.Contains(elseBranch, koolTargetVar) {
		t.Fatalf("else branch must register %s; branch=%q", koolTargetVar, elseBranch)
	}
	if !strings.Contains(elseBranch, "to targetDir") {
		t.Fatalf("else branch must assign targetDir to %s; branch=%q", koolTargetVar, elseBranch)
	}
	match := reuseScriptMatchBranch(s)
	if match == "" {
		t.Fatalf("missing match branch: %q", s)
	}
	if strings.Contains(match, koolTargetVar) {
		t.Fatalf("match branch must not register %s: %q", koolTargetVar, match)
	}
}
```