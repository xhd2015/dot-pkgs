## Expected

- `DefaultPublishPort` is `23891`.
- Type constants:
  - `TypeSeatalkMessageReceived` == `"seatalk.message.received"`
  - `TypeSeatalkSessionOpened` == `"seatalk.session.opened"`
  - `TypeAgentTTYStarted` == `"agent.tty.started"`
- Source constants:
  - `SourceSeatalkLocalBot` == `"seatalk.local-bot"`
  - `SourceAgentRun` == `"agent-run"`

## Side Effects

- None.

## Errors

- `err` is nil.

## Exit Code

- Success.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("constants read: unexpected error: %v", err)
	}
	if resp.DefaultPublishPort != 23891 {
		t.Fatalf("DefaultPublishPort: got %d, want 23891", resp.DefaultPublishPort)
	}
	checks := []struct {
		name, got, want string
	}{
		{"TypeSeatalkMessageReceived", resp.TypeSeatalkMessageReceived, "seatalk.message.received"},
		{"TypeSeatalkSessionOpened", resp.TypeSeatalkSessionOpened, "seatalk.session.opened"},
		{"TypeAgentTTYStarted", resp.TypeAgentTTYStarted, "agent.tty.started"},
		{"SourceSeatalkLocalBot", resp.SourceSeatalkLocalBot, "seatalk.local-bot"},
		{"SourceAgentRun", resp.SourceAgentRun, "agent-run"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Fatalf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
}
```
