package install

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestResolveLatestStableURLStopsBeforeZipBody(t *testing.T) {
	const zipBytes = 8 << 20 // 8 MiB would be fatal if drained
	payload := make([]byte, zipBytes)
	var zipBodyReads atomic.Int64

	mux := http.NewServeMux()
	mux.HandleFunc("/stable/latest", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/downloads/iTerm2-3_6_11.zip", http.StatusFound)
	})
	mux.HandleFunc("/downloads/iTerm2-3_6_11.zip", func(w http.ResponseWriter, r *http.Request) {
		n, _ := w.Write(payload)
		zipBodyReads.Add(int64(n))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url, ver, err := ResolveLatestStableURL(ctx, ResolveOpts{
		LatestURL:  srv.URL + "/stable/latest",
		HTTPClient: srv.Client(), // nil CheckRedirect → production stop-before-zip
	})
	if err != nil {
		t.Fatal(err)
	}
	if ver != "3.6.11" {
		t.Fatalf("version=%q want 3.6.11", ver)
	}
	if !strings.HasSuffix(url, "/downloads/iTerm2-3_6_11.zip") {
		t.Fatalf("url=%q", url)
	}
	if got := zipBodyReads.Load(); got != 0 {
		t.Fatalf("zip body reads=%d want 0 (resolve must not download archive)", got)
	}
}

func TestResolveLatestStableURLRelativeLocation(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "iTerm2-1_2_3.zip")
		w.WriteHeader(http.StatusFound)
	})
	mux.HandleFunc("/iTerm2-1_2_3.zip", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("zip handler must not be hit")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	url, ver, err := ResolveLatestStableURL(context.Background(), ResolveOpts{
		LatestURL:  srv.URL + "/latest",
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if ver != "1.2.3" {
		t.Fatalf("version=%q", ver)
	}
	if !strings.HasSuffix(url, "/iTerm2-1_2_3.zip") {
		t.Fatalf("url=%q", url)
	}
}

func TestResolveLatestStableURLHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, _, err := ResolveLatestStableURL(context.Background(), ResolveOpts{
		LatestURL:  srv.URL,
		HTTPClient: srv.Client(),
	})
	if err == nil {
		t.Fatal("want error")
	}
}

func TestResolveLatestStableURLNonZipRedirect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/not-a-zip", http.StatusFound)
	})
	mux.HandleFunc("/not-a-zip", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "html")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	_, _, err := ResolveLatestStableURL(context.Background(), ResolveOpts{
		LatestURL:  srv.URL + "/latest",
		HTTPClient: srv.Client(),
	})
	if err == nil {
		t.Fatal("want non-zip error")
	}
}
