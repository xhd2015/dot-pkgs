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
			if opts.URL != api.BillingURL {
				return nil, fmt.Errorf("unexpected url %q", opts.URL)
			}
			if opts.Auth.AccessToken != "access-token-fixture" {
				return nil, fmt.Errorf("unexpected token")
			}
			return []byte(fixtureBillingLimited), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if snap.UsedPercent != 73 || snap.Source != "billing" || snap.Email != "user@example.com" {
		t.Fatalf("snap = %+v", snap)
	}
	if ensureCalls != 1 || getCalls != 1 {
		t.Fatalf("ensure=%d get=%d", ensureCalls, getCalls)
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
			if gets == 1 {
				return nil, fmt.Errorf("grok api: HTTP 401 from %s", api.BillingURL)
			}
			if opts.Auth.AccessToken != "refreshed-access" {
				return nil, fmt.Errorf("expected refreshed token")
			}
			return []byte(fixtureBillingLimited), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Used != 73 {
		t.Fatalf("snap = %+v", snap)
	}
	if gets != 2 || len(ensureForce) != 2 || ensureForce[1] != true {
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
			return nil, fmt.Errorf("grok api: HTTP 500 from %s", api.BillingURL)
		},
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("err = %v", err)
	}
}
