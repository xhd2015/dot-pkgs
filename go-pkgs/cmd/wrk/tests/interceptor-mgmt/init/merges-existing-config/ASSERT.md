## Expected

- Exit code 0.
- Interceptor stub present (`enabled: false`).
- Top-level `notes` key still `"keep-me"`.

## Side Effects

- Merge write; does not wipe unknown keys.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertMgmtNeutralStub(t, req.WrkHome)

	var root map[string]json.RawMessage
	if err := json.Unmarshal(readMgmtConfigBytes(t, req.WrkHome), &root); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	notesRaw, ok := root["notes"]
	if !ok {
		t.Fatal("expected top-level notes key preserved")
	}
	var notes string
	if err := json.Unmarshal(notesRaw, &notes); err != nil {
		t.Fatalf("notes: %v", err)
	}
	if notes != "keep-me" {
		t.Fatalf("notes: want keep-me, got %q", notes)
	}
}
```
