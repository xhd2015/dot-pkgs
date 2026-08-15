# Scenario

**Feature**: one host produces two ingress rules — host service then 404 catch-all

```
# BuildConfigFromState
1 host app.example.com -> http://127.0.0.1:6321
  -> Ingress len 2
  -> [0] hostname+service
  -> [1] empty hostname, http_status:404
```

## Preconditions

- StateIn has TunnelID, CredentialsFile, and exactly one Hosts entry.

## Steps

1. Build StateIn with single host fixture.

## Context

- Requirement scenarios 8 and 10 (single-host form of 404 last).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "single-host-plus-404")
	req.StateIn = mustState(
		"team-shared",
		"id-single-host",
		"/tmp/cred/single.json",
		map[string]*cloudflare.HostEntry{
			"app.example.com": {Service: "http://127.0.0.1:6321"},
		},
	)
	return nil
}
```
