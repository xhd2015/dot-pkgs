## Expected

- `resp.ExitCode` is 0.
- `resp.Stdout` is **repo-level** help: mentions `list` and that
  `repo list --help` shows list options; ends with `\n`.
- `resp.Stdout` is **not** list leaf help (must not document list-only flags
  such as `--search-description`).
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
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHelpStdout(t, resp, "list")
	lower := strings.ToLower(resp.Stdout)
	// Must point users at list leaf help for list options.
	if !strings.Contains(lower, "repo list --help") && !strings.Contains(lower, "list --help") {
		t.Fatalf("repo-level help should mention list --help for list options, got stdout=%q", resp.Stdout)
	}
	// Distinguish from list leaf help (listHelp documents --search-description).
	assertNotContainsFold(t, resp.Stdout, "--search-description")
}
```
