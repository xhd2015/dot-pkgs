# shell/grok/usage — Grok account billing via shell/grok/api

## Version

0.0.2

Classic TDD doctests for package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/grok/usage`.

**Default layer: L2** in-process library API with injectable `Ensure` / `GetJSON`.
No real network, no real `~/.grok/auth.json`. Parallel-safe.

**Out of scope:** Marcus wiring, agent-pro PTY fallback, live cli-chat-proxy e2e,
Cloudflare `grok.com/rest/rate-limits` (request/token windows).

Fixtures use generic identities (`user@example.com`) — no personal hosts in
scenario copy beyond public product URLs used as API constants.

## DSN (Domain Specific Notion)

### Participants

- **`Fetch(ctx, opts)`** — load/ensure auth → GET monthly `api.BillingURL` and
  weekly `api.BillingCreditsURL` → `NormalizeJSON` each → `SelectPreferred`.
  On HTTP 401/403 for a GET, force-refresh once and retry that GET once.
- **`NormalizeJSON(raw, source)`** — map monthly `used`/`monthlyLimit` and/or
  credits `creditUsagePercent` / `currentPeriod` into `Snapshot`.
- **`SelectPreferred`** — monthly wins when `MonthlyLimit > 0`; else weekly when
  `UsedPercent >= 0`; else monthly uncapped; else weekly.
- **`Snapshot`** — `Used`, `MonthlyLimit`, `UsedPercent` / `RemainingPercent`
  (`-1` when unknown), `ResetAt`, `PeriodType` (`monthly`/`weekly`), `Source`
  (`billing`), `Email`.
- **`FetchOpts.ForceRefresh`** — always refresh access token before GETs.

### Behaviors

**Fetch**

- Monthly capped + weekly present → prefer monthly percents / `PeriodType=monthly`
- Monthly uncapped + weekly credits → prefer weekly percents / `PeriodType=weekly`
- Monthly uncapped only (credits same shape / no percent) → percents `-1`
- `ForceRefresh=true` → Ensure called with ForceRefresh
- GET 401 then OK after forced refresh → success
- Both GETs fail → error

**Normalize**

- Covered primarily by L1 unit tests; L2 leaves exercise Fetch end-to-end.

## Tree

```
fetch/
├── billing-ok/                 # monthly capped wins over weekly
├── monthly-open-weekly/        # monthlyLimit=0 → weekly credits %
├── unlimited/                  # monthly uncapped, no weekly % → unknown %
├── force-refresh/              # ForceRefresh passed to Ensure
├── unauthorized-retry/         # 401 → force refresh → retry OK
└── http-fail/                  # both GETs fail
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

const fixtureBillingCreditsWeekly = `{
  "config": {
    "billingPeriodStart": "2026-08-28T00:55:25.179446+00:00",
    "billingPeriodEnd": "2026-09-04T00:55:25.179446+00:00",
    "creditUsagePercent": 2.0,
    "currentPeriod": {
      "start": "2026-08-28T00:55:25.179446+00:00",
      "end": "2026-09-04T00:55:25.179446+00:00",
      "type": "USAGE_PERIOD_TYPE_WEEKLY"
    },
    "productUsage": [{"product": "GrokBuild", "usagePercent": 2.0}]
  }
}`

// Request is filled root→leaf.
type Request struct {
	Operation string // fetch

	// FetchMode: billing-ok | monthly-open-weekly | unlimited | force-refresh |
	// unauthorized-retry | http-fail
	FetchMode string
}

// Response observes Fetch outputs.
type Response struct {
	Used             int64
	MonthlyLimit     int64
	RemainingPercent int
	UsedPercent      int
	PeriodType       string
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

func bodyForURL(mode, url string, gets int) ([]byte, error) {
	switch mode {
	case "billing-ok":
		switch url {
		case api.BillingURL:
			return []byte(fixtureBillingLimited), nil
		case api.BillingCreditsURL:
			return []byte(fixtureBillingCreditsWeekly), nil
		}
	case "monthly-open-weekly":
		switch url {
		case api.BillingURL:
			return []byte(fixtureBillingUnlimited), nil
		case api.BillingCreditsURL:
			return []byte(fixtureBillingCreditsWeekly), nil
		}
	case "unlimited", "force-refresh":
		return []byte(fixtureBillingUnlimited), nil
	case "unauthorized-retry":
		if gets == 1 && url == api.BillingURL {
			return nil, fmt.Errorf("grok api: HTTP 401 from %s", api.BillingURL)
		}
		switch url {
		case api.BillingURL:
			return []byte(fixtureBillingLimited), nil
		case api.BillingCreditsURL:
			return []byte(fixtureBillingCreditsWeekly), nil
		}
	case "http-fail":
		return nil, fmt.Errorf("grok api: HTTP 500 from %s", url)
	}
	return nil, fmt.Errorf("unknown FetchMode/url %q %q", mode, url)
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
		if req.FetchMode == "unauthorized-retry" && gets > 1 && opts.Auth.AccessToken != "refreshed-access" {
			return nil, fmt.Errorf("expected refreshed token, got %q", opts.Auth.AccessToken)
		}
		return bodyForURL(req.FetchMode, opts.URL, gets)
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
	resp.PeriodType = snap.PeriodType
	resp.Source = snap.Source
	resp.Email = snap.Email
	if !snap.ResetAt.IsZero() {
		resp.ResetUnix = snap.ResetAt.Unix()
	}
	return resp, nil
}
```
