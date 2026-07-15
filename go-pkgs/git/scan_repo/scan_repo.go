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
	ListRemotes        bool
	ListWorktrees      bool
	Stderr             io.Writer
	// OnRepo is invoked immediately when a repository is discovered during the walk.
	OnRepo func(Repo) error

	// CacheRoot is the root of the on-disk mirror cache. Empty means the
	// product default ($HOME/.cache/git-repo-scan) when cache is enabled.
	// Tests always pass an explicit temp dir. Write side effects are
	// controlled with NoCache.
	CacheRoot string
	// NoCache, when true, disables cache read and write for this Scan.
	NoCache bool
	// Refresh, when true and cache is enabled, forces a cold full walk +
	// mirror rewrite even when the root is warm-eligible.
	Refresh bool

	// WarmRefreshBudget bounds wall time spent rewalking refresh units on a
	// warm scan (P4). 0 means the product default (1s). Negative means no
	// refresh work. Not applied on cold full walks.
	WarmRefreshBudget time.Duration
	// YoungAge is the minimum age (now - refreshed_at) before a refresh unit
	// is eligible. 0 means the product default (60s).
	YoungAge time.Duration
	// Now, if non-nil, is the clock for age/budget decisions; nil means time.Now.
	// Tests prefer stamped refreshed_at + YoungAge/Budget over real sleeps.
	Now func() time.Time
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

	// NoCache=true disables all mirror I/O. When cache is on, empty CacheRoot
	// resolves to the product default ($HOME/.cache/git-repo-scan).
	cacheRoot := resolveCacheRoot(opts)

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

		// Warm path: complete root cache → serve is_repo marks + liveness, no full re-walk.
		// NoCache leaves cacheRoot empty. Refresh forces cold full walk + rewrite.
		useWarm := cacheRoot != "" && !opts.Refresh && rootWarmEligible(cacheRoot, absRoot)

		if opts.OnRepo != nil {
			var walkErr error
			handleOnRepo := func(repo Repo) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				repo = enrichRepo(ctx, repo, opts)
				if onErr := opts.OnRepo(repo); onErr != nil {
					repo.Error = appendRepoError(repo.Error, onErr.Error())
				}
				result.Repos = append(result.Repos, repo)
				return nil
			}
			if useWarm {
				// Collect served first so budgeted refresh can dedupe against them.
				var served []Repo
				served, walkErr = warmServeRoot(ctx, absRoot, cacheRoot, nil)
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
					_, walkErr = warmBudgetRefresh(ctx, absRoot, opts, cacheRoot, ignore, served, handleOnRepo)
				}
			} else {
				_, walkErr = walkRoot(ctx, absRoot, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr, handleOnRepo, cacheRoot)
			}
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
			found, walkErr = warmServeRoot(ctx, absRoot, cacheRoot, nil)
			if walkErr == nil {
				found, walkErr = warmBudgetRefresh(ctx, absRoot, opts, cacheRoot, ignore, found, nil)
			}
		} else {
			found, walkErr = walkRoot(ctx, absRoot, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr, nil, cacheRoot)
		}
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
		result.Repos = append(result.Repos, found...)
	}

	sort.Slice(result.Repos, func(i, j int) bool {
		return result.Repos[i].Path < result.Repos[j].Path
	})

	if opts.OnRepo == nil {
		for i := range result.Repos {
			result.Repos[i] = enrichRepo(ctx, result.Repos[i], opts)
		}
	}

	return result, nil
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
