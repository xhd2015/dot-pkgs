## Expected

- Save + Load succeed (`err == nil`).
- Loaded `TunnelName`, `TunnelID`, `CredentialsFile` match StateIn.
- Hosts map has exactly one entry for `app.example.com` with same Service
  and OwnerPID.

## Side Effects

- Creates `state.json` (and parent dirs) under a temp tunnel dir.

## Errors

- None.

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("save_load error: %v", err)
	}
	if resp == nil || resp.State == nil {
		t.Fatal("nil response or State")
	}
	want := req.StateIn
	got := resp.State
	if got.TunnelName != want.TunnelName {
		t.Fatalf("TunnelName = %q, want %q", got.TunnelName, want.TunnelName)
	}
	if got.TunnelID != want.TunnelID {
		t.Fatalf("TunnelID = %q, want %q", got.TunnelID, want.TunnelID)
	}
	if got.CredentialsFile != want.CredentialsFile {
		t.Fatalf("CredentialsFile = %q, want %q", got.CredentialsFile, want.CredentialsFile)
	}
	assertHostsEqual(t, got.Hosts, want.Hosts)
}
```
