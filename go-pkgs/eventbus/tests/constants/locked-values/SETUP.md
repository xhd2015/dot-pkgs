# Scenario

**Feature**: DefaultPublishPort and v1 type/source strings match locked contract

```
# locked wire vocabulary
DefaultPublishPort -> 23891
TypeSeatalkMessageReceived -> "seatalk.message.received"
TypeSeatalkSessionOpened -> "seatalk.session.opened"
TypeAgentTTYStarted -> "agent.tty.started"
TypeAgentTTYRestarted -> "agent.tty.restarted"
ReasonTTYNew -> "new"
ReasonTTYFollowup -> "followup"
ReasonTTYResume -> "resume"
SourceSeatalkLocalBot -> "seatalk.local-bot"
SourceAgentRun -> "agent-run"
```

## Steps

1. Ensure `req.Op` is `"constants"` so `Run` reads package constants into `Response`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "constants"
	return nil
}
```
