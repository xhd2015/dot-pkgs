package seed

import "strings"

// wellKnownHosts are VCS hosts where cmd/go derives the repo URL without
// go-import meta discovery (same class as github.com in the go command).
var wellKnownHosts = map[string]int{
	// host → path elements after host that form the repo root
	"github.com":    2,
	"bitbucket.org": 2,
	"gitlab.com":    2,
}

// VCSRootHTTPS returns the https:// VCS root URL for a module path when the
// host is well-known. ok is false for vanity / unknown hosts (caller should skip).
func VCSRootHTTPS(modulePath string) (url string, ok bool) {
	modulePath = strings.TrimSpace(modulePath)
	modulePath = strings.TrimPrefix(modulePath, "https://")
	modulePath = strings.TrimPrefix(modulePath, "http://")
	parts := strings.Split(modulePath, "/")
	if len(parts) < 2 {
		return "", false
	}
	host := parts[0]
	n, known := wellKnownHosts[host]
	if !known {
		return "", false
	}
	if len(parts) < 1+n {
		return "", false
	}
	root := strings.Join(parts[:1+n], "/")
	return "https://" + root, true
}

// IsWellKnown reports whether modulePath uses a well-known VCS host.
func IsWellKnown(modulePath string) bool {
	_, ok := VCSRootHTTPS(modulePath)
	return ok
}
