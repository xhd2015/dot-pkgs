package release

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchLatestReleaseTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/xhd2015/myapp/releases/latest" {
			w.Header().Set("Location", "https://github.com/xhd2015/myapp/releases/tag/v1.2.3")
			w.WriteHeader(http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	old := defaultBaseURL
	defaultBaseURL = srv.URL
	defer func() { defaultBaseURL = old }()

	got, err := FetchLatestReleaseTag(context.Background(), "xhd2015", "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "v1.2.3" {
		t.Fatalf("got %q, want %q", got, "v1.2.3")
	}
}

func TestFetchLatestReleaseVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://github.com/xhd2015/myapp/releases/tag/v3.0.1")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	old := defaultBaseURL
	defaultBaseURL = srv.URL
	defer func() { defaultBaseURL = old }()

	got, err := FetchLatestReleaseVersion(context.Background(), "xhd2015", "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "3.0.1" {
		t.Fatalf("got %q, want %q", got, "3.0.1")
	}
}

func TestFetchLatestReleaseTag_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	old := defaultBaseURL
	defaultBaseURL = srv.URL
	defer func() { defaultBaseURL = old }()

	_, err := FetchLatestReleaseTag(context.Background(), "xhd2015", "norelease")
	if err == nil {
		t.Fatal("expected error for non-302 response")
	}
}

func TestFetchLatestReleaseTag_MissingLocation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	old := defaultBaseURL
	defaultBaseURL = srv.URL
	defer func() { defaultBaseURL = old }()

	_, err := FetchLatestReleaseTag(context.Background(), "xhd2015", "myapp")
	if err == nil {
		t.Fatal("expected error for missing Location header")
	}
}

func TestFetchLatestReleaseTag_EmptyOwnerRepo(t *testing.T) {
	_, err := FetchLatestReleaseTag(context.Background(), "", "myapp")
	if err == nil {
		t.Fatal("expected error for empty owner")
	}
	_, err = FetchLatestReleaseTag(context.Background(), "xhd2015", "")
	if err == nil {
		t.Fatal("expected error for empty repo")
	}
}
