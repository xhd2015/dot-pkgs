package githook

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ExcludeRepos pattern syntax:
//
//   - origin:PATTERN  match only the normalized origin URL (host/path)
//   - dir:PATTERN     match only the repo top-level directory
//   - PATTERN         match both
//
// PATTERN is a glob where "*" matches any run of characters (including "/")
// and "?" matches exactly one character; every other character is literal.
// Origin matching ignores case; dir matching is case-sensitive. A pattern
// without "/" also matches the origin host part and the dir base name. A
// leading "~" in a dir pattern expands to the user home directory.

// NormalizeOriginURL reduces a git remote URL to "host/path" so patterns see
// one uniform form: scheme, userinfo, port, and a trailing ".git" are
// stripped, scp-style "host:owner/repo" becomes "host/owner/repo", and the
// host is lowercased. Local paths and unparseable remotes return "".
func NormalizeOriginURL(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ""
	}
	if u, err := url.Parse(remote); err == nil && u.Host != "" {
		return joinOriginHostPath(strings.ToLower(u.Hostname()), u.Path)
	}
	if strings.Contains(remote, "://") {
		return ""
	}
	hostPart := remote
	if at := strings.LastIndex(hostPart, "@"); at >= 0 {
		hostPart = hostPart[at+1:]
	}
	if colon := strings.Index(hostPart, ":"); colon >= 0 {
		return joinOriginHostPath(strings.ToLower(strings.Trim(hostPart[:colon], "[]")), hostPart[colon+1:])
	}
	if strings.Contains(hostPart, "/") {
		return ""
	}
	return strings.ToLower(strings.Trim(hostPart, "[]"))
}

func joinOriginHostPath(host string, path string) string {
	host = strings.Trim(host, "[]")
	path = strings.TrimSuffix(path, ".git")
	path = strings.Trim(path, "/")
	if path == "" {
		return host
	}
	return host + "/" + path
}

// repoExcluded reports whether the current repo matches any ExcludeRepos
// pattern, resolving the origin URL and the repo top-level directory via git.
func (f DomainFilter) repoExcluded() (bool, error) {
	var origin string
	remote, ok, err := GitOptionalOutput("config", "--get", "remote.origin.url")
	if err != nil {
		return false, err
	}
	if ok {
		origin = NormalizeOriginURL(strings.TrimSpace(remote))
	}
	dir, err := repoToplevelDir()
	if err != nil {
		return false, err
	}
	return MatchRepoPatterns(f.ExcludeRepos, origin, dir)
}

// repoToplevelDir returns the current repo top-level directory, falling back
// to the working directory when git cannot report one.
func repoToplevelDir() (string, error) {
	out, ok, err := GitOptionalOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	if ok {
		if dir := strings.TrimSpace(out); dir != "" {
			return dir, nil
		}
	}
	return os.Getwd()
}

// MatchRepoPatterns reports whether any pattern excludes the repo described by
// origin (a normalized "host/path" from NormalizeOriginURL) and dir (an
// absolute repo path). Patterns are OR'ed. An empty pattern is an error.
func MatchRepoPatterns(patterns []string, origin string, dir string) (bool, error) {
	for _, raw := range patterns {
		kind, pattern := cutRepoPatternTarget(strings.TrimSpace(raw))
		if strings.TrimSpace(pattern) == "" {
			return false, repoPatternEmptyErr(kind)
		}
		if kind != "dir" && matchRepoOrigin(pattern, origin) {
			return true, nil
		}
		if kind != "origin" {
			matched, err := matchRepoDir(pattern, dir)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
	}
	return false, nil
}

func repoPatternEmptyErr(kind string) error {
	if kind == "" {
		return fmt.Errorf("invalid --exclude-repo: pattern must not be empty")
	}
	return fmt.Errorf("invalid --exclude-repo: %s pattern must not be empty", kind)
}

// cutRepoPatternTarget splits an optional "origin:"/"dir:" target prefix from
// the glob pattern; bare patterns target both origin and dir.
func cutRepoPatternTarget(pattern string) (kind string, rest string) {
	lower := strings.ToLower(pattern)
	if strings.HasPrefix(lower, "origin:") {
		return "origin", pattern[len("origin:"):]
	}
	if strings.HasPrefix(lower, "dir:") {
		return "dir", pattern[len("dir:"):]
	}
	return "", pattern
}

// normalizeRepoPatterns trims ExcludeRepos values and rejects empty patterns
// at parse time.
func normalizeRepoPatterns(patterns []string) ([]string, error) {
	cleaned := make([]string, 0, len(patterns))
	for _, raw := range patterns {
		pattern := strings.TrimSpace(raw)
		kind, rest := cutRepoPatternTarget(pattern)
		if strings.TrimSpace(rest) == "" {
			return nil, repoPatternEmptyErr(kind)
		}
		cleaned = append(cleaned, pattern)
	}
	return cleaned, nil
}

// matchRepoOrigin reports whether the pattern matches the normalized origin
// URL, or its host part when the pattern has no "/". Matching ignores case.
func matchRepoOrigin(pattern string, origin string) bool {
	if origin == "" {
		return false
	}
	if repoGlobRegexp(pattern, true).MatchString(origin) {
		return true
	}
	if !strings.Contains(pattern, "/") {
		if host, _, ok := strings.Cut(origin, "/"); ok && repoGlobRegexp(pattern, true).MatchString(host) {
			return true
		}
	}
	return false
}

// matchRepoDir reports whether the pattern matches the repo dir, or its base
// name when the pattern has no "/". Matching is case-sensitive.
func matchRepoDir(pattern string, dir string) (bool, error) {
	if dir == "" {
		return false, nil
	}
	pattern, err := expandTildePrefix(pattern)
	if err != nil {
		return false, err
	}
	pattern = strings.TrimRight(pattern, "/")
	if pattern == "" {
		return false, nil
	}
	re := repoGlobRegexp(pattern, false)
	for _, candidate := range repoDirCandidates(dir) {
		if re.MatchString(candidate) {
			return true, nil
		}
	}
	if !strings.Contains(pattern, "/") && re.MatchString(filepath.Base(filepath.Clean(dir))) {
		return true, nil
	}
	return false, nil
}

// repoDirCandidates lists the spellings of dir that dir patterns may target:
// the path as given and, when different, its symlink-resolved form.
func repoDirCandidates(dir string) []string {
	cleaned := filepath.Clean(dir)
	candidates := []string{cleaned}
	if resolved, err := filepath.EvalSymlinks(cleaned); err == nil && resolved != cleaned {
		candidates = append(candidates, resolved)
	}
	return candidates
}

// expandTildePrefix expands a leading "~" or "~/" to the user home directory;
// "~user" is left untouched.
func expandTildePrefix(pattern string) (string, error) {
	if pattern != "~" && !strings.HasPrefix(pattern, "~/") {
		return pattern, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expand ~ in --exclude-repo pattern: %w", err)
	}
	return home + strings.TrimPrefix(pattern, "~"), nil
}

// repoGlobRegexp compiles pattern into an anchored regexp where "*" matches
// any run of characters including "/" and "?" matches exactly one character;
// every other character is literal.
func repoGlobRegexp(pattern string, ignoreCase bool) *regexp.Regexp {
	var b strings.Builder
	if ignoreCase {
		b.WriteString("(?i)")
	}
	b.WriteString("^")
	for _, r := range pattern {
		switch r {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteString(".")
		default:
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}
