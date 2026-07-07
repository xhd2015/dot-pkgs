## Expected Output

Summary contains `codex sessions:` and `codex skills:`.
Summary omits `grok sessions:`, `grok projects:`, and `grok skills:`.

## Expected

- `SummaryText` includes codex topic lines.
- `SummaryText` excludes all grok topic lines.

## Errors

- `err` is nil.
- Missing codex lines or unexpected grok lines.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	text := resp.SummaryText

	for _, needle := range []string{
		"analyse-files summary",
		"codex sessions:",
		"codex skills:",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("summary missing %q:\n%s", needle, text)
		}
	}

	for _, needle := range []string{
		"grok sessions:",
		"grok projects:",
		"grok skills:",
	} {
		if strings.Contains(text, needle) {
			t.Fatalf("summary unexpectedly contains %q:\n%s", needle, text)
		}
	}
}
```