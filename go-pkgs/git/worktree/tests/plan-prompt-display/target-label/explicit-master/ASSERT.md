## Expected

- Prompt uses `master` in question and `# master: fast forward`.
- Must not contain `merge into main?` or `# main: fast forward`.
- Output does **not** start with a leading blank line (`"\n"`).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(resp.Output, "\n") {
		t.Fatal(`FormatPlanPrompt output must not start with leading "\n"`)
	}
	firstLine, _, _ := strings.Cut(resp.Output, "\n")
	if firstLine != "branch feature is ahead, merge into master?" {
		t.Fatalf("first line = %q, want branch feature is ahead, merge into master?", firstLine)
	}
	if strings.Contains(resp.Output, "merge into main?") || strings.Contains(resp.Output, "# main: fast forward") {
		t.Fatalf("output must use master label, got:\n%s", resp.Output)
	}
	// Template must not begin with a raw-string newline (that would expect want:"" line 1).
	assert.Output(t, resp.Output, `<contains>
branch feature is ahead, merge into master?
  # master: fast forward
</contains>`)
}
```
