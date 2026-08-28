package usage

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/grok/api"
)

func TestFetch_BillingOK(t *testing.T) {
	var ensureCalls, getCalls int
	var urls []string
	snap, err := Fetch(context.Background(), FetchOpts{
		Auth: api.Auth{
			AccessToken: "access-token-fixture",
			Email:       "user@example.com",
			ExpiresAt:   time.Now().Add(time.Hour),
		},
		Ensure: func(ctx context.Context, opts api.EnsureOpts) (api.Auth, error) {
			ensureCalls++
			return opts.Auth, nil
		},
		GetJSON: func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
			getCalls++
			urls = append(urls, opts.URL)
			if opts.Auth.AccessToken != "access-token-fixture" {
				return nil, fmt.Errorf("unexpected token")
			}
			switch opts.URL {
			case api.BillingURL:
				return []byte(fixtureBillingLimited), nil
			case api.BillingCreditsURL:
				return []byte(fixtureBillingCreditsWeekly), nil
			default:
				return nil, fmt.Errorf("unexpected url %q", opts.URL)
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Monthly cap wins over weekly credits.
	if snap.UsedPercent != 73 || snap.PeriodType != PeriodMonthly || snap.Source != "billing" || snap.Email != "user@example.com" {
		t.Fatalf("snap = %+v", snap)
	}
	if ensureCalls != 1 || getCalls != 2 {
		t.Fatalf("ensure=%d get=%d urls=%v", ensureCalls, getCalls, urls)
	}
}

func TestFetch_MonthlyUncappedUsesWeekly(t *testing.T) {
	snap, err := Fetch(context.Background(), FetchOpts{
		Auth: api.Auth{
			AccessToken: "access-token-fixture",
			Email:       "user@example.com",
			ExpiresAt:   time.Now().Add(time.Hour),
		},
		Ensure: func(ctx context.Context, opts api.EnsureOpts) (api.Auth, error) {
			return opts.Auth, nil
		},
		GetJSON: func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
			switch opts.URL {
			case api.BillingURL:
				return []byte(fixtureBillingUnlimited), nil
			case api.BillingCreditsURL:
				return []byte(fixtureBillingCreditsWeekly), nil
			default:
				return nil, fmt.Errorf("unexpected url %q", opts.URL)
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if snap.UsedPercent != 2 || snap.PeriodType != PeriodWeekly {
		t.Fatalf("snap = %+v", snap)
	}
}

func TestFetch_ForceRefreshPassed(t *testing.T) {
	var forced bool
	_, err := Fetch(context.Background(), FetchOpts{
		Auth: api.Auth{
			AccessToken:  "access-token-fixture",
			RefreshToken: "refresh-token-fixture",
			ExpiresAt:    time.Now().Add(time.Hour),
		},
		ForceRefresh: true,
		Ensure: func(ctx context.Context, opts api.EnsureOpts) (api.Auth, error) {
			forced = opts.ForceRefresh
			return opts.Auth, nil
		},
		GetJSON: func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
			return []byte(fixtureBillingUnlimited), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !forced {
		t.Fatal("want ForceRefresh passed to Ensure")
	}
}

func TestFetch_UnauthorizedRetriesOnce(t *testing.T) {
	var ensureForce []bool
	var gets int
	snap, err := Fetch(context.Background(), FetchOpts{
		Auth: api.Auth{
			AccessToken:  "access-token-fixture",
			RefreshToken: "refresh-token-fixture",
			ExpiresAt:    time.Now().Add(time.Hour),
		},
		Ensure: func(ctx context.Context, opts api.EnsureOpts) (api.Auth, error) {
			ensureForce = append(ensureForce, opts.ForceRefresh)
			out := opts.Auth
			if opts.ForceRefresh {
				out.AccessToken = "refreshed-access"
			}
			return out, nil
		},
		GetJSON: func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
			gets++
			// First monthly GET fails auth; retry + credits succeed.
			if gets == 1 {
				return nil, fmt.Errorf("grok api: HTTP 401 from %s", api.BillingURL)
			}
			if opts.Auth.AccessToken != "refreshed-access" {
				return nil, fmt.Errorf("expected refreshed token")
			}
			switch opts.URL {
			case api.BillingURL:
				return []byte(fixtureBillingLimited), nil
			case api.BillingCreditsURL:
				return []byte(fixtureBillingCreditsWeekly), nil
			default:
				return nil, fmt.Errorf("unexpected url %q", opts.URL)
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Used != 73 || snap.UsedPercent != 73 {
		t.Fatalf("snap = %+v", snap)
	}
	// initial Ensure + forced refresh on 401
	if gets < 2 || len(ensureForce) < 2 || ensureForce[1] != true {
		t.Fatalf("gets=%d ensureForce=%v", gets, ensureForce)
	}
}

func TestFetch_BothFail(t *testing.T) {
	_, err := Fetch(context.Background(), FetchOpts{
		Auth: api.Auth{AccessToken: "access-token-fixture", ExpiresAt: time.Now().Add(time.Hour)},
		Ensure: func(ctx context.Context, opts api.EnsureOpts) (api.Auth, error) {
			return opts.Auth, nil
		},
		GetJSON: func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
			return nil, fmt.Errorf("grok api: HTTP 500 from %s", opts.URL)
		},
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("err = %v", err)
	}
}
