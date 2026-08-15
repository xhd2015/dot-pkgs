## Expected

- `resp.ExitCode` is 0.
- `resp.Stderr` is empty.
- `resp.Stdout` parses as JSON array with 2 elements.
- First element `FullName` is `alice/alpha`, `matched_by` is `["owned"]`.
- Second element `FullName` is `alice/beta`, `matched_by` is `["owned"]`.

## Side Effects

- Mock `gh api user` and `gh repo list alice` invoked.

## Errors

- Harness `err` is nil.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", resp.Stderr)
	}
	var results []struct {
		FullName  string   `json:"FullName"`
		MatchedBy []string `json:"matched_by"`
	}
	if unmarshalErr := json.Unmarshal([]byte(resp.Stdout), &results); unmarshalErr != nil {
		t.Fatalf("invalid JSON stdout: %v\n%q", unmarshalErr, resp.Stdout)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 JSON results, got %d: %s", len(results), resp.Stdout)
	}
	if results[0].FullName != "alice/alpha" || results[1].FullName != "alice/beta" {
		t.Fatalf("unexpected sort order: %+v", results)
	}
	for _, r := range results {
		if len(r.MatchedBy) != 1 || r.MatchedBy[0] != "owned" {
			t.Fatalf("expected matched_by [owned], got %+v", r.MatchedBy)
		}
	}
	if !strings.Contains(resp.Stdout, "\n") {
		t.Fatalf("expected indented JSON with newlines, got %q", resp.Stdout)
	}
}
```