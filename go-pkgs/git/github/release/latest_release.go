package release

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// defaultBaseURL is overridable in tests (e.g. to point at httptest.NewServer).
var defaultBaseURL = "https://github.com"

// defaultTimeout is the per-request timeout for the 302 redirect fetch.
const defaultTimeout = 10 * time.Second

// FetchLatestReleaseTag returns the latest release tag name (e.g. "v1.0.13")
// for owner/repo by following the releases/latest 302 redirect. No GitHub API
// token is required; this uses the web redirect, not the REST API, so it is
// not subject to the anonymous API rate limit.
func FetchLatestReleaseTag(ctx context.Context, owner, repo string) (string, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return "", fmt.Errorf("owner and repo are required")
	}

	baseURL := strings.TrimRight(defaultBaseURL, "/")
	url := fmt.Sprintf("%s/%s/%s/releases/latest", baseURL, owner, repo)

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("expected 302 from %s, got %d", url, resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("missing Location header from %s", url)
	}

	tag := location
	if idx := strings.LastIndex(location, "/"); idx >= 0 {
		tag = location[idx+1:]
	}
	tag = strings.TrimSpace(tag)
	// Strip URL fragments / query that should not appear but be safe.
	if idx := strings.IndexAny(tag, "?#"); idx >= 0 {
		tag = tag[:idx]
	}
	if tag == "" {
		return "", fmt.Errorf("could not parse tag from Location: %s", location)
	}
	return tag, nil
}

// FetchLatestReleaseVersion is like FetchLatestReleaseTag but strips the "v"
// prefix, returning e.g. "1.0.13".
func FetchLatestReleaseVersion(ctx context.Context, owner, repo string) (string, error) {
	tag, err := FetchLatestReleaseTag(ctx, owner, repo)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(tag, "v"), nil
}
