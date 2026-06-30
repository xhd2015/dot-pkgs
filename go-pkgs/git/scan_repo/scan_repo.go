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

	Remotes   []Remote
	Worktrees []Worktree
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
}

var defaultIgnoreDirs = []string{
	".git", "node_modules", "vendor", ".venv", "__pycache__", "dist", "build", "target",
}

func Scan(ctx context.Context, opts Options) ([]Repo, error) {
	if len(opts.Roots) == 0 {
		return nil, fmt.Errorf("at least one root is required")
	}

	ignore, err := buildIgnoreConfig(opts)
	if err != nil {
		return nil, err
	}

	var repos []Repo
	for _, root := range opts.Roots {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		absRoot, err := validateRoot(root)
		if err != nil {
			return nil, err
		}

		found, err := walkRoot(ctx, absRoot, opts.MaxDepth, ignore, opts.Verbose, opts.Stderr)
		if err != nil {
			return nil, err
		}
		repos = append(repos, found...)
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Path < repos[j].Path
	})

	for i := range repos {
		if opts.ListRemotes {
			remotes, err := listRemotes(ctx, repos[i].Path)
			if err != nil {
				return nil, err
			}
			repos[i].Remotes = remotes
		}
		if opts.ListWorktrees && repos[i].RepoType == RepoTypeMain {
			worktrees, err := listWorktrees(ctx, repos[i].Path)
			if err != nil {
				return nil, err
			}
			repos[i].Worktrees = worktrees
		}
	}

	return repos, nil
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