package scan_repo

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
}

var defaultIgnoreDirs = []string{
	".git", "node_modules", "vendor", ".venv", "__pycache__", "dist", "build", "target",
}

func Scan(ctx context.Context, opts Options) (Result, error) {
	if len(opts.Roots) == 0 {
		return Result{}, fmt.Errorf("at least one root is required")
	}

	ignore, err := buildIgnoreConfig(opts)
	if err != nil {
		return Result{}, err
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

		if opts.OnRepo != nil {
			_, walkErr := walkRoot(ctx, absRoot, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr, func(repo Repo) error {
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
			})
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

		found, walkErr := walkRoot(ctx, absRoot, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr, nil)
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
