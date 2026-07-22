package scan_repo

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type RepoType string

const (
	RepoTypeMain     RepoType = "main"
	RepoTypeWorktree RepoType = "worktree"
)

type Repo struct {
	Path     string
	Name     string
	GitDir   string
	RepoType RepoType
	Error    string `json:"error,omitempty"`

	Remotes   []Remote
	Worktrees []Worktree
}

type RootError struct {
	Root  string
	Error string
}

type Result struct {
	Repos      []Repo
	RootErrors []RootError
}

type Remote struct {
	Name  string
	URL   string
	Host  string
	Owner string
	Repo  string
}

type Worktree struct {
	Path   string
	Head   string
	IsMain bool
}

type Options struct {
	Roots              []string
	MaxDepth           int
	IgnoreDirs         []string
	IgnoreDirBasenames []string
	Verbose            bool
	// Debug, when true, writes greppable phase-level "scan:" lines to Stderr
	// (default os.Stderr): cache root, per-root mode=warm|cold + reason,
	// warm serve candidates/live/duration, refresh summary, root total.
	// Orthogonal to Verbose (permission/remote skip warnings).
	Debug         bool
	ListRemotes   bool
	ListWorktrees bool
	Stderr        io.Writer
	// OnRepo is invoked immediately when a repository is discovered during the walk.
	OnRepo func(Repo) error

	// CacheRoot is the root of the on-disk durable cache (repo index + walk log).
	// Empty means the product default ($HOME/.cache/git-repo-scan) when cache is
	// enabled. Dense mirror under CacheRoot/mirror is retired and never written.
	// Tests always pass an explicit temp dir. Write side effects are controlled
	// with NoCache.
	CacheRoot string
	// NoCache, when true, disables cache read and write for this Scan.
	NoCache bool
	// Refresh, when true and cache is enabled, forces a cold full walk +
	// index/walk rewrite even when the root is warm-eligible.
	Refresh bool

	// WarmRefreshBudget bounds wall time spent rewalking refresh units on a
	// warm scan (P4). 0 means the product default (1s). Negative means no
	// refresh work. Not applied on cold full walks.
	WarmRefreshBudget time.Duration
	// YoungAge is the minimum age (now - unit_dir.ModTime) before a refresh unit
	// is eligible. 0 means the product default (60s).
	YoungAge time.Duration
	// Now, if non-nil, is the clock for age/budget decisions; nil means time.Now.
	// Tests prefer stamped unit ModTime + YoungAge/Budget over real sleeps.
	Now func() time.Time

	// LastScanEnd is the prior scan completion time used to select the
	// walk-log consume sync budget (P4). Zero means read home/meta.json
	// last_scan_end when present; if still unknown, budget treats delta as
	// ancient (full 1s tier).
	LastScanEnd time.Time
}

// debugf writes a greppable "scan: …" line to opts.Stderr when Debug is on.
// No-op when Debug is false (must not emit the "scan:" substring).
func debugf(opts Options, format string, args ...interface{}) {
	if !opts.Debug {
		return
	}
	w := stderrWriter(opts.Stderr)
	fmt.Fprintf(w, "scan: "+format+"\n", args...)
}

// defaultCacheRoot is the product cache store when cache is enabled and
// Options.CacheRoot is empty: $HOME/.cache/git-repo-scan.
const defaultCacheDirName = "git-repo-scan"

var defaultIgnoreDirs = []string{
	".git", "node_modules", "vendor", ".venv", "__pycache__", "dist", "build", "target",
}

// resolveCacheRoot returns the on-disk cache root for this Scan.
// Empty when NoCache; otherwise opts.CacheRoot or the product default.
func resolveCacheRoot(opts Options) string {
	if opts.NoCache {
		return ""
	}
	if opts.CacheRoot != "" {
		return opts.CacheRoot
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		// Fallback for environments without a home directory.
		if home = os.Getenv("HOME"); home == "" {
			return ""
		}
	}
	return filepath.Join(home, ".cache", defaultCacheDirName)
}

func Scan(ctx context.Context, opts Options) (Result, error) {
	if len(opts.Roots) == 0 {
		return Result{}, fmt.Errorf("at least one root is required")
	}

	ignore, err := buildIgnoreConfig(opts)
	if err != nil {
		return Result{}, err
	}

	// NoCache=true disables all durable cache I/O (index + walk log). When cache
	// is on, empty CacheRoot resolves to the product default.
	cacheRoot := resolveCacheRoot(opts)
	// Never walk into the cache store itself (e.g. bare scan of $HOME with
	// cache under ~/.cache/git-repo-scan).
	if cacheRoot != "" {
		if ignore.fullPaths == nil {
			ignore.fullPaths = make(map[string]struct{})
		}
		ignore.fullPaths[filepath.Clean(cacheRoot)] = struct{}{}
	}

	if opts.Debug {
		if cacheRoot == "" {
			debugf(opts, "cacheRoot=<none>")
		} else {
			debugf(opts, "cacheRoot=%s", cacheRoot)
		}
	}

	var result Result
	for _, root := range opts.Roots {
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		default:
		}

		absRoot, err := validateRoot(root)
		if err != nil {
			result.RootErrors = append(result.RootErrors, RootError{
				Root:  root,
				Error: err.Error(),
			})
			continue
		}

		// Warm path: usable home/repos.json under root → serve index + liveness, no full re-walk.
		// NoCache leaves cacheRoot empty. Refresh forces cold full walk + rewrite.
		useWarm, modeReason := rootCacheMode(cacheRoot, absRoot, opts)
		rootStart := time.Now()
		if useWarm {
			debugf(opts, "root=%s mode=warm reason=%s", absRoot, modeReason)
		} else {
			debugf(opts, "root=%s mode=cold reason=%s", absRoot, modeReason)
		}

		if opts.OnRepo != nil {
			var walkErr error
			var rootCollected []Repo
			handleOnRepo := func(repo Repo) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				// Resolve (remotes/worktrees) then base-path filter before emit.
				repo = enrichRepo(ctx, repo, opts)
				filtered := filterReposUnderRoot(absRoot, []Repo{repo})
				if len(filtered) == 0 {
					return nil
				}
				repo = filtered[0]
				if onErr := opts.OnRepo(repo); onErr != nil {
					repo.Error = appendRepoError(repo.Error, onErr.Error())
				}
				rootCollected = append(rootCollected, repo)
				result.Repos = append(result.Repos, repo)
				return nil
			}
			if useWarm {
				// Collect served first so budgeted refresh can dedupe against them.
				var served []Repo
				var serveStats warmServeStats
				served, serveStats, walkErr = warmServeRootOpts(ctx, absRoot, cacheRoot, opts, nil)
				debugf(opts, "serve root=%s candidates=%d live=%d duration=%s",
					absRoot, serveStats.candidates, serveStats.live, serveStats.duration)
				if walkErr == nil {
					for _, repo := range served {
						if err := handleOnRepo(repo); err != nil {
							walkErr = err
							break
						}
					}
				}
				if walkErr == nil {
					// New units merge via handleOnRepo; existing seeds path dedupe.
					var refreshStats warmRefreshStats
					_, refreshStats, walkErr = warmBudgetRefresh(ctx, absRoot, opts, cacheRoot, ignore, served, handleOnRepo)
					debugf(opts, "refresh root=%s budget=%s eligible=%d refreshed=%d duration=%s",
						absRoot, refreshStats.budget, refreshStats.eligible, refreshStats.refreshed, refreshStats.duration)
				}
				// P4: walk-log consume (re-list visits, gone/new, gen_end G+1).
				if walkErr == nil {
					_, walkErr = consumeWalkLog(ctx, cacheRoot, opts, absRoot, rootCollected, handleOnRepo)
				}
			} else {
				_, walkErr = walkRoot(ctx, absRoot, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr, handleOnRepo, cacheRoot, true)
				// P2: cold walk seeds durable home/repos.json.
				if walkErr == nil {
					if seedErr := seedHomeRepoIndex(cacheRoot, absRoot, rootCollected); seedErr != nil {
						walkErr = seedErr
					}
				}
				// P3: seal cold generation (gen_end gen=1) + walk cursor at EOF.
				if walkErr == nil {
					if sealErr := SealColdWalkGenEnd(cacheRoot, 1); sealErr != nil {
						walkErr = sealErr
					}
				}
			}
			debugf(opts, "root=%s total duration=%s", absRoot, time.Since(rootStart))
			if walkErr != nil {
				if walkErr == ctx.Err() {
					return Result{}, walkErr
				}
				result.RootErrors = append(result.RootErrors, RootError{
					Root:  root,
					Error: walkErr.Error(),
				})
			}
			continue
		}

		var found []Repo
		var walkErr error
		if useWarm {
			var serveStats warmServeStats
			found, serveStats, walkErr = warmServeRootOpts(ctx, absRoot, cacheRoot, opts, nil)
			debugf(opts, "serve root=%s candidates=%d live=%d duration=%s",
				absRoot, serveStats.candidates, serveStats.live, serveStats.duration)
			if walkErr == nil {
				var refreshStats warmRefreshStats
				found, refreshStats, walkErr = warmBudgetRefresh(ctx, absRoot, opts, cacheRoot, ignore, found, nil)
				debugf(opts, "refresh root=%s budget=%s eligible=%d refreshed=%d duration=%s",
					absRoot, refreshStats.budget, refreshStats.eligible, refreshStats.refreshed, refreshStats.duration)
			}
			// P4: walk-log consume (re-list visits, gone/new, gen_end G+1).
			if walkErr == nil {
				found, walkErr = consumeWalkLog(ctx, cacheRoot, opts, absRoot, found, nil)
			}
		} else {
			found, walkErr = walkRoot(ctx, absRoot, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr, nil, cacheRoot, true)
			// P2: cold walk seeds durable home/repos.json.
			if walkErr == nil {
				if seedErr := seedHomeRepoIndex(cacheRoot, absRoot, found); seedErr != nil {
					walkErr = seedErr
				}
			}
			// P3: seal cold generation (gen_end gen=1) + walk cursor at EOF.
			if walkErr == nil {
				if sealErr := SealColdWalkGenEnd(cacheRoot, 1); sealErr != nil {
					walkErr = sealErr
				}
			}
		}
		debugf(opts, "root=%s total duration=%s", absRoot, time.Since(rootStart))
		if walkErr != nil {
			if walkErr == ctx.Err() {
				return Result{}, walkErr
			}
			result.RootErrors = append(result.RootErrors, RootError{
				Root:  root,
				Error: walkErr.Error(),
			})
			continue
		}
		// Resolve worktrees/remotes, then drop paths outside this scan root.
		for i := range found {
			found[i] = enrichRepo(ctx, found[i], opts)
		}
		found = filterReposUnderRoot(absRoot, found)
		result.Repos = append(result.Repos, found...)
	}

	sort.Slice(result.Repos, func(i, j int) bool {
		return result.Repos[i].Path < result.Repos[j].Path
	})

	return result, nil
}

// filterReposUnderRoot drops top-level repos whose Path is not under absRoot
// and strips Worktrees entries whose Path is not under absRoot.
// Call after enrichRepo when ListWorktrees may have filled Worktrees.
func filterReposUnderRoot(absRoot string, repos []Repo) []Repo {
	if len(repos) == 0 {
		return repos
	}
	out := make([]Repo, 0, len(repos))
	for _, r := range repos {
		if !pathIsUnderRoot(absRoot, r.Path) {
			continue
		}
		if n := len(r.Worktrees); n > 0 {
			kept := make([]Worktree, 0, n)
			for _, wt := range r.Worktrees {
				if pathIsUnderRoot(absRoot, wt.Path) {
					kept = append(kept, wt)
				}
			}
			if len(kept) == 0 {
				r.Worktrees = nil
			} else {
				r.Worktrees = kept
			}
		}
		out = append(out, r)
	}
	return out
}

func enrichRepo(ctx context.Context, repo Repo, opts Options) Repo {
	if repo.Error != "" {
		return repo
	}
	if opts.ListRemotes {
		remotes, err := listRemotes(ctx, repo.Path)
		if err != nil {
			repo.Error = appendRepoError(repo.Error, err.Error())
			return repo
		}
		repo.Remotes = remotes
	}
	if opts.ListWorktrees && repo.RepoType == RepoTypeMain {
		worktrees, err := listWorktrees(ctx, repo.Path)
		if err != nil {
			repo.Error = appendRepoError(repo.Error, err.Error())
			return repo
		}
		repo.Worktrees = worktrees
	}
	return repo
}

func appendRepoError(existing, msg string) string {
	if msg == "" {
		return existing
	}
	if existing == "" {
		return msg
	}
	return existing + "; " + msg
}

type ignoreConfig struct {
	basenames map[string]struct{}
	fullPaths map[string]struct{}
}

func buildIgnoreConfig(opts Options) (ignoreConfig, error) {
	cfg := ignoreConfig{
		basenames: make(map[string]struct{}, len(defaultIgnoreDirs)+len(opts.IgnoreDirBasenames)),
		fullPaths: make(map[string]struct{}, len(opts.IgnoreDirs)),
	}
	for _, name := range defaultIgnoreDirs {
		cfg.basenames[name] = struct{}{}
	}
	for _, name := range opts.IgnoreDirBasenames {
		cfg.basenames[name] = struct{}{}
	}
	for _, dir := range opts.IgnoreDirs {
		norm, err := normalizeIgnoreDir(dir)
		if err != nil {
			return ignoreConfig{}, fmt.Errorf("%s: %w", dir, err)
		}
		cfg.fullPaths[norm] = struct{}{}
	}
	return cfg, nil
}

func normalizeIgnoreDir(path string) (string, error) {
	expanded, err := expandPath(path)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func expandPath(path string) (string, error) {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return home, nil
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func validateRoot(root string) (string, error) {
	expanded, err := expandPath(root)
	if err != nil {
		return "", fmt.Errorf("%s: %w", root, err)
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("%s: %w", root, err)
	}
	abs = filepath.Clean(abs)

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("%s: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s: not a directory", root)
	}
	return abs, nil
}
