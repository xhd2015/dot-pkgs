# Scenario

**Feature**: SaveTunnelState then LoadTunnelState round-trips one host entry

```
# save + load
StateIn with 1 host
  -> SaveTunnelState(dir, StateIn)
  -> LoadTunnelState(dir)
  -> equal tunnel metadata + Hosts
```

## Preconditions

- Mode `save_load`.
- StateIn has TunnelName, TunnelID, CredentialsFile, and one HostEntry.

## Steps

1. Set Mode and a single-host StateIn fixture.

## Context

- Requirement scenario 6: round-trip equal for 1 host.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "save-load-roundtrip")
	req.Mode = "save_load"
	req.StateIn = mustState(
		"team-shared",
		"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		"/tmp/cred/team-shared.json",
		map[string]*cloudflare.HostEntry{
			"app.example.com": {
				Service:  "http://127.0.0.1:6321",
				OwnerPID: 4242,
			},
		},
	)
	return nil
}
```
