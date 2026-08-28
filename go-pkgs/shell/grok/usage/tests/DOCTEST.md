# shell/grok/usage — Grok account billing via shell/grok/api

## Version

0.0.1

Classic TDD doctests for package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/grok/usage`.

**Default layer: L2** in-process library API with injectable `Ensure` / `GetJSON`.
No real network, no real `~/.grok/auth.json`. Parallel-safe.

**Out of scope:** Marcus wiring, agent-pro PTY fallback, live cli-chat-proxy e2e,
Cloudflare `grok.com/rest/rate-limits`.

Fixtures use generic identities (`user@example.com`) — no personal hosts in
scenario copy beyond public product URLs used as API constants.

## DSN (Domain Specific Notion)

### Participants

- **`Fetch(ctx, opts)`** — load/ensure auth → GET `api.BillingURL` →
  `NormalizeJSON`. On HTTP 401/403, force-refresh once and retry GET once.
- **`NormalizeJSON(raw, source)`** — map `config.used` / `monthlyLimit` /
  billing period into `Snapshot`.
- **`Snapshot`** — `Used`, `MonthlyLimit`, `UsedPercent` / `RemainingPercent`
  (`-1` when limit is 0), `ResetAt`, `Source` (`billing`), `Email`.
- **`FetchOpts.ForceRefresh`** — always refresh access token before GET.

### Behaviors

**Fetch**

- Billing OK with limit → `Source=billing`, percents from used/limit
- Billing OK with `monthlyLimit=0` → percents `-1`
- `ForceRefresh=true` → Ensure called with ForceRefresh
- GET 401 then OK after forced refresh → success
- Non-auth GET failure → error

**Normalize**

- Covered primarily by L1 unit tests; L2 leaves exercise Fetch end-to-end.

## Tree

```
fetch/
├── billing-ok/            # limited billing succeeds
├── unlimited/             # monthlyLimit=0 → unknown %
├── force-refresh/         # ForceRefresh passed to Ensure
├── unauthorized-retry/    # 401 → force refresh → retry OK
└── http-fail/             # non-auth GET failure
```

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/grok/usage/tests
doctest test ./shell/grok/usage/tests
```

## Types + Run + helpers

```go
import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/grok/api"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/grok/usage"
)

const fixtureBillingLimited = `{
  "config": {
    "monthlyLimit": {"val": 100},
    "used": {"val": 73},
    "billingPeriodStart": "2026-08-01T00:00:00+00:00",
    "billingPeriodEnd": "2026-09-01T00:00:00+00:00"
  }
}`

const fixtureBillingUnlimited = `{
  "config": {
    "monthlyLimit": {"val": 0},
    "used": {"val": 73},
    "billingPeriodStart": "2026-08-01T00:00:00+00:00",
    "billingPeriodEnd": "2026-09-01T00:00:00+00:00"
  }
}`

// Request is filled root→leaf.
type Request struct {
	Operation string // fetch

	// FetchMode: billing-ok | unlimited | force-refresh | unauthorized-retry | http-fail
	FetchMode string
}

// Response observes Fetch outputs.
type Response struct {
	Used             int64
	MonthlyLimit     int64
	RemainingPercent int
	UsedPercent      int
	Source           string
	Email            string
	ResetUnix        int64
	EnsureForced     []bool
	GetURLs          []string
	GetCount         int
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
}

func assertEqual(t *testing.T, name string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %#v want %#v", name, got, want)
	}
}

func assertContains(t *testing.T, name, got, substr string) {
	t.Helper()
	if !strings.Contains(got, substr) {
		t.Fatalf("%s: %q does not contain %q", name, got, substr)
	}
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.Operation == "" {
		t.Fatal("Operation not set")
	}
	if req.Operation != "fetch" {
		t.Fatalf("unknown Operation %q", req.Operation)
	}

	var ensureForced []bool
	var urls []string
	var gets int

	ensure := func(ctx context.Context, opts api.EnsureOpts) (api.Auth, error) {
		ensureForced = append(ensureForced, opts.ForceRefresh)
		out := opts.Auth
		if opts.ForceRefresh {
			out.AccessToken = "refreshed-access"
		}
		return out, nil
	}

	get := func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
		gets++
		urls = append(urls, opts.URL)
		switch req.FetchMode {
		case "billing-ok":
			return []byte(fixtureBillingLimited), nil
		case "unlimited":
			return []byte(fixtureBillingUnlimited), nil
		case "force-refresh":
			return []byte(fixtureBillingUnlimited), nil
		case "unauthorized-retry":
			if gets == 1 {
				return nil, fmt.Errorf("grok api: HTTP 401 from %s", api.BillingURL)
			}
			if opts.Auth.AccessToken != "refreshed-access" {
				return nil, fmt.Errorf("expected refreshed token, got %q", opts.Auth.AccessToken)
			}
			return []byte(fixtureBillingLimited), nil
		case "http-fail":
			return nil, fmt.Errorf("grok api: HTTP 500 from %s", api.BillingURL)
		default:
			return nil, fmt.Errorf("unknown FetchMode %q", req.FetchMode)
		}
	}

	force := req.FetchMode == "force-refresh"
	snap, err := usage.Fetch(context.Background(), usage.FetchOpts{
		Auth: api.Auth{
			AccessToken:  "access-token-fixture",
			RefreshToken: "refresh-token-fixture",
			Email:        "user@example.com",
			ExpiresAt:    time.Now().Add(time.Hour),
		},
		ForceRefresh: force,
		Ensure:       ensure,
		GetJSON:      get,
	})
	resp := &Response{
		EnsureForced: ensureForced,
		GetURLs:      urls,
		GetCount:     gets,
	}
	if err != nil {
		return resp, err
	}
	resp.Used = snap.Used
	resp.MonthlyLimit = snap.MonthlyLimit
	resp.RemainingPercent = snap.RemainingPercent
	resp.UsedPercent = snap.UsedPercent
	resp.Source = snap.Source
	resp.Email = snap.Email
	if !snap.ResetAt.IsZero() {
		resp.ResetUnix = snap.ResetAt.Unix()
	}
	return resp, nil
}
```
