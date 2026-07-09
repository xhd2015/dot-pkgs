## Expected

- `resp.ExitCode` is 0.
- `resp.Stdout` is **repo-level** help (same class as bare `repo`): mentions
  `list` and list `--help`; not list leaf help; trailing `\n`.
- `resp.Stderr` is empty.

## Side Effects

- No `gh` invocation.

## Errors

- `err` from harness is nil.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHelpStdout(t, resp, "list")
	lower := strings.ToLower(resp.Stdout)
	if !strings.Contains(lower, "repo list --help") && !strings.Contains(lower, "list --help") {
		t.Fatalf("repo-level help should mention list --help for list options, got stdout=%q", resp.Stdout)
	}
	assertNotContainsFold(t, resp.Stdout, "--search-description")
}
```
