package scan_repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// Resolve reasons returned by ResolveByOrigin.
const (
	ReasonUnderCWD    = "under-cwd"
	ReasonContainsCWD = "contains-cwd"
	ReasonMain        = "main"
	ReasonWorktree    = "worktree"
)

// OriginIdentity is a normalized remote identity (host + owner + repo).
type OriginIdentity struct {
	Host  string
	Owner string
	Repo  string
}

// ResolveByOriginResult is the chosen checkout for an origin URL.
type ResolveByOriginResult struct {
	Path   string
	Reason string
	Repo   Repo
}

// ParseOriginIdentity normalizes a git remote URL (or owner/repo) into host/owner/repo.
// Bare "owner/repo" assumes github.com.
func ParseOriginIdentity(raw string) (OriginIdentity, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return OriginIdentity{}, fmt.Errorf("origin URL is required")
	}

	if !strings.Contains(raw, "://") && !strings.Contains(raw, "@") && strings.Count(raw, "/") == 1 {
		parts := strings.Split(raw, "/")
		owner := strings.TrimSpace(parts[0])
		repo := strings.TrimSuffix(strings.TrimSpace(parts[1]), ".git")
		if owner == "" || repo == "" {
			return OriginIdentity{}, fmt.Errorf("invalid origin %q", raw)
		}
		return OriginIdentity{Host: "github.com", Owner: owner, Repo: repo}, nil
	}

	host := remoteHost(raw)
	owner, repo, ok := ParseRemoteOwnerRepo(raw)
	if !ok || host == "" || owner == "" || repo == "" {
		return OriginIdentity{}, fmt.Errorf("cannot parse origin identity from %q", raw)
	}
	return OriginIdentity{
		Host:  strings.ToLower(host),
		Owner: owner,
		Repo:  repo,
	}, nil
}

// RemotesMatchOrigin reports whether any remote matches the identity (host/owner/repo).
// Prefer origin name when present, but any matching remote counts.
func RemotesMatchOrigin(remotes []Remote, id OriginIdentity) bool {
	id.Host = strings.ToLower(strings.TrimSpace(id.Host))
	id.Owner = strings.TrimSpace(id.Owner)
	id.Repo = strings.TrimSpace(id.Repo)
	if id.Host == "" || id.Owner == "" || id.Repo == "" {
		return false
	}
	if matchNamedRemote(remotes, "origin", id) {
		return true
	}
	for _, remote := range remotes {
		if remote.Name == "origin" {
			continue
		}
		if remoteMatchesIdentity(remote, id) {
			return true
		}
	}
	return false
}

func matchNamedRemote(remotes []Remote, name string, id OriginIdentity) bool {
	for _, remote := range remotes {
		if remote.Name != name {
			continue
		}
		return remoteMatchesIdentity(remote, id)
	}
	return false
}

func remoteMatchesIdentity(remote Remote, id OriginIdentity) bool {
	host := strings.ToLower(strings.TrimSpace(remote.Host))
	if host == "" {
		host = remoteHost(remote.URL)
	}
	return host == id.Host &&
		strings.EqualFold(remote.Owner, id.Owner) &&
		strings.EqualFold(remote.Repo, id.Repo)
}

// DefaultResolveRoots returns scan roots: $HOME plus the git toplevel of cwd when
// present and distinct from $HOME.
func DefaultResolveRoots(cwd string) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home: %w", err)
	}
	home = filepath.Clean(home)
	roots := []string{home}

	cwd = strings.TrimSpace(cwd)
	if cwd == "" {
		return roots, nil
	}
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return roots, nil
	}
	toplevel, err := worktree.ShowToplevel(absCwd)
	if err != nil || strings.TrimSpace(toplevel) == "" {
		return roots, nil
	}
	toplevel = filepath.Clean(toplevel)
	if toplevel != home {
		roots = append(roots, toplevel)
	}
	return roots, nil
}

// ResolveByOrigin scans opts.Roots (via Scan, honoring cache options) for checkouts
// whose remotes match originURL, then picks one using preferUnder ranking:
// under-cwd → contains-cwd → main → worktree.
//
// ListRemotes is forced on so remotes are available for matching. preferUnder is
// typically the caller's cwd; empty skips cwd-based ranking.
func ResolveByOrigin(ctx context.Context, opts Options, originURL string, preferUnder string) (*ResolveByOriginResult, error) {
	id, err := ParseOriginIdentity(originURL)
	if err != nil {
		return nil, err
	}
	if len(opts.Roots) == 0 {
		return nil, fmt.Errorf("at least one root is required")
	}

	opts.ListRemotes = true
	result, err := Scan(ctx, opts)
	if err != nil {
		return nil, err
	}

	candidates := collectOriginCandidates(result.Repos, id)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no local checkout matching origin %s/%s/%s", id.Host, id.Owner, id.Repo)
	}

	preferUnder = strings.TrimSpace(preferUnder)
	if preferUnder != "" {
		if abs, absErr := filepath.Abs(preferUnder); absErr == nil {
			preferUnder = filepath.Clean(abs)
		} else {
			preferUnder = filepath.Clean(preferUnder)
		}
	}

	chosen := pickOriginCandidate(candidates, preferUnder)
	return &ResolveByOriginResult{
		Path:   chosen.Path,
		Reason: chosen.reason,
		Repo:   chosen.Repo,
	}, nil
}

type rankedCandidate struct {
	Repo
	reason   string
	priority int
}

func collectOriginCandidates(repos []Repo, id OriginIdentity) []Repo {
	seen := make(map[string]struct{})
	var out []Repo

	add := func(r Repo) {
		path := filepath.Clean(r.Path)
		if path == "" || path == "." {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		r.Path = path
		out = append(out, r)
	}

	for _, r := range repos {
		if strings.TrimSpace(r.Error) != "" {
			continue
		}
		if !RemotesMatchOrigin(r.Remotes, id) {
			continue
		}
		add(r)
		if r.RepoType != RepoTypeMain {
			continue
		}
		for _, wt := range r.Worktrees {
			if wt.IsMain || strings.TrimSpace(wt.Path) == "" {
				continue
			}
			add(Repo{
				Path:     wt.Path,
				Name:     filepath.Base(wt.Path),
				GitDir:   r.GitDir,
				RepoType: RepoTypeWorktree,
				Remotes:  r.Remotes,
			})
		}
	}
	return out
}

func pickOriginCandidate(candidates []Repo, preferUnder string) rankedCandidate {
	ranked := make([]rankedCandidate, 0, len(candidates))
	for _, c := range candidates {
		priority, reason := originRank(c.Path, c.RepoType, preferUnder)
		ranked = append(ranked, rankedCandidate{Repo: c, reason: reason, priority: priority})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].priority != ranked[j].priority {
			return ranked[i].priority < ranked[j].priority
		}
		// under-cwd: deepest (longest) path; contains-cwd: closest parent (shortest)
		switch ranked[i].priority {
		case 0:
			if len(ranked[i].Path) != len(ranked[j].Path) {
				return len(ranked[i].Path) > len(ranked[j].Path)
			}
		case 1:
			if len(ranked[i].Path) != len(ranked[j].Path) {
				return len(ranked[i].Path) < len(ranked[j].Path)
			}
		}
		return ranked[i].Path < ranked[j].Path
	})
	return ranked[0]
}

func originRank(path string, repoType RepoType, preferUnder string) (priority int, reason string) {
	path = filepath.Clean(path)
	if preferUnder != "" {
		if path != preferUnder && pathIsUnderRoot(preferUnder, path) {
			return 0, ReasonUnderCWD
		}
		if pathIsUnderRoot(path, preferUnder) {
			return 1, ReasonContainsCWD
		}
	}
	if repoType == RepoTypeMain {
		return 2, ReasonMain
	}
	return 3, ReasonWorktree
}
