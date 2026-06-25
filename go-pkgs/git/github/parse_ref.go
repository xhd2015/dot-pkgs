package github

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseRef extracts GitHub owner and repository name from a match-project github
// field (owner/repo slug, https URL, or git@github.com:owner/repo.git).
func ParseRef(ref string) (owner, name string, err error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", fmt.Errorf("empty github ref")
	}

	if strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "http://") {
		u, parseErr := url.Parse(ref)
		if parseErr != nil {
			return "", "", fmt.Errorf("parse github url: %w", parseErr)
		}
		if !isGitHubHost(u.Hostname()) {
			return "", "", fmt.Errorf("not a github.com url: %q", ref)
		}
		return ownerRepoFromURLPath(u.Path)
	}

	if strings.HasPrefix(strings.ToLower(ref), "github.com/") {
		return ownerRepoFromURLPath("/" + strings.TrimPrefix(ref, "github.com/"))
	}

	if o, n, ok := parseSCPStyleGitHub(ref); ok {
		return o, n, nil
	}

	if strings.Count(ref, "/") == 1 && !strings.ContainsAny(ref, " \t@:") {
		return ownerRepoFromURLPath("/" + ref)
	}

	return "", "", fmt.Errorf("invalid github ref: %q", ref)
}

func isGitHubHost(host string) bool {
	return strings.EqualFold(strings.TrimSpace(host), "github.com")
}

func ownerRepoFromURLPath(path string) (owner, name string, err error) {
	path = strings.Trim(path, "/")
	if path == "" {
		return "", "", fmt.Errorf("missing owner/repo in github ref")
	}
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("missing owner/repo in github ref")
	}
	owner = parts[len(parts)-2]
	name = strings.TrimSuffix(parts[len(parts)-1], ".git")
	if owner == "" || name == "" {
		return "", "", fmt.Errorf("missing owner/repo in github ref")
	}
	return owner, name, nil
}

func parseSCPStyleGitHub(raw string) (owner, name string, ok bool) {
	hostPart := raw
	if at := strings.LastIndex(hostPart, "@"); at >= 0 {
		hostPart = hostPart[at+1:]
	}
	colon := strings.Index(hostPart, ":")
	if colon < 0 {
		return "", "", false
	}
	host := strings.ToLower(hostPart[:colon])
	if !isGitHubHost(host) {
		return "", "", false
	}
	o, n, err := ownerRepoFromURLPath("/" + hostPart[colon+1:])
	if err != nil {
		return "", "", false
	}
	return o, n, true
}