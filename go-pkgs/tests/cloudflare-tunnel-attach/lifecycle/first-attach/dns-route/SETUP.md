# Scenario

**Feature**: first attach invokes DNS route for the hostname via CommandRunner

```
# Attach
Attach(a.example.com)
  -> runner.Exec includes tunnel route dns … a.example.com
```

## Preconditions

- Same first-attach fixtures as registry-and-config.
- Fake runner records all Exec argument slices.

## Steps

1. Attach once with Domain `a.example.com`.
2. Assert inspects fake runner call log for route/dns/hostname tokens.

## Context

- Requirement scenario 6: DNS route called with hostname.
- Production shape (from existing RouteDNS):  
  `cloudflared tunnel route dns --overwrite-dns <tunnel> <hostname>`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "dns-route")
	req.Domain = "a.example.com"
	req.LocalURL = "http://127.0.0.1:6321"
	req.Sequence = nil
	return nil
}
```
