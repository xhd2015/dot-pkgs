package usage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/api"
)

const fixtureSpendJSON = `{
  "plan_type": "business",
  "email": "user@example.com",
  "rate_limit": null,
  "spend_control": {
    "individual_limit": {
      "used_percent": 62,
      "remaining_percent": 38,
      "reset_at": 1788220800
    }
  }
}`

func TestFetch_CodexUsageOK(t *testing.T) {
	var hits []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits = append(hits, r.URL.Path)
		_, _ = w.Write([]byte(fixtureSpendJSON))
	}))
	defer srv.Close()

	snap, err := Fetch(context.Background(), FetchOpts{
		Auth: api.Auth{
			AccessToken: "access-token-fixture",
			AccountID:   "00000000-0000-4000-8000-000000000001",
		},
		GetJSON: func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
			opts.URL = srv.URL + "/backend-api/codex/usage"
			opts.HTTPClient = srv.Client()
			return api.GetJSON(ctx, opts)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if snap.RemainingPercent != 38 || snap.Source != "codex/usage" {
		t.Fatalf("snap = %+v", snap)
	}
	if len(hits) != 1 {
		t.Fatalf("hits = %v", hits)
	}
}

func TestFetch_FallbackWham(t *testing.T) {
	var urls []string
	snap, err := Fetch(context.Background(), FetchOpts{
		Auth: api.Auth{
			AccessToken: "access-token-fixture",
			AccountID:   "00000000-0000-4000-8000-000000000001",
		},
		GetJSON: func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
			urls = append(urls, opts.URL)
			if strings.Contains(opts.URL, "codex/usage") {
				return nil, fmt.Errorf("codex api: HTTP 403 from usage endpoint")
			}
			if strings.Contains(opts.URL, "wham/usage") {
				return []byte(fixtureSpendJSON), nil
			}
			return nil, fmt.Errorf("unexpected url %q", opts.URL)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Source != "wham/usage" || snap.RemainingPercent != 38 {
		t.Fatalf("snap = %+v", snap)
	}
	if len(urls) != 2 {
		t.Fatalf("urls = %v", urls)
	}
}

func TestFetch_BothFail(t *testing.T) {
	_, err := Fetch(context.Background(), FetchOpts{
		Auth: api.Auth{
			AccessToken: "access-token-fixture",
			AccountID:   "00000000-0000-4000-8000-000000000001",
		},
		GetJSON: func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
			return nil, fmt.Errorf("codex api: HTTP 403 from usage endpoint")
		},
	})
	if err == nil {
		t.Fatal("want error")
	}
}
