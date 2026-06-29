package scan_repo

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var errFound = errors.New("scan_repo: github match found")

// FindLocalMainByGitHub walks opts.Roots in order and returns the first main
// checkout whose github.com remote matches owner/repo. Worktree rows are skipped.
// Walk stops at the first match.
func FindLocalMainByGitHub(ctx context.Context, opts Options, owner, repo string) (*Repo, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}
	if len(opts.Roots) == 0 {
		return nil, fmt.Errorf("at least one root is required")
	}

	ignore, err := buildIgnoreConfig(opts)
	if err != nil {
		return nil, err
	}

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

		found, err := walkRootFindGitHub(ctx, absRoot, owner, repo, opts.MaxDepth, ignore, opts.Verbose)
		if err == errFound {
			return found, nil
		}
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("no local main repo for github.com/%s/%s", owner, repo)
}

func walkRootFindGitHub(ctx context.Context, root, owner, repo string, maxDepth int, ignore ignoreConfig, verbose bool) (*Repo, error) {
	var found *Repo

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if isPermissionError(walkErr) {
				maybeWarnSkip(verbose, path, walkErr)
				return filepath.SkipDir
			}
			return walkErr
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !d.IsDir() {
			return nil
		}

		if path != root {
			cleanPath := filepath.Clean(path)
			if _, skip := ignore.fullPaths[cleanPath]; skip {
				return filepath.SkipDir
			}
			if _, skip := ignore.basenames[d.Name()]; skip {
				return filepath.SkipDir
			}
			if maxDepth > 0 && depthFromRoot(root, path) > maxDepth {
				return filepath.SkipDir
			}
		}

		gitPath := filepath.Join(path, ".git")
		info, statErr := os.Stat(gitPath)
		if statErr != nil || !(info.IsDir() || info.Mode().IsRegular()) {
			return nil
		}

		gitDir, repoType, resolveErr := resolveGitDir(path, gitPath, info)
		if resolveErr != nil {
			return resolveErr
		}
		if repoType != RepoTypeMain {
			return filepath.SkipDir
		}

		remotes, listErr := listRemotes(ctx, path)
		if listErr != nil {
			return listErr
		}
		if !remotesMatchGitHub(remotes, owner, repo) {
			return filepath.SkipDir
		}

		found = &Repo{
			Path:     path,
			Name:     filepath.Base(path),
			GitDir:   gitDir,
			RepoType: RepoTypeMain,
			Remotes:  remotes,
		}
		return errFound
	})
	if err == errFound {
		return found, errFound
	}
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func remotesMatchGitHub(remotes []Remote, owner, repo string) bool {
	if matchGitHubRemote(remotes, "origin", owner, repo) {
		return true
	}
	for _, remote := range remotes {
		if remote.Name == "origin" {
			continue
		}
		if remote.Host == "github.com" && remote.Owner == owner && remote.Repo == repo {
			return true
		}
	}
	return false
}

func matchGitHubRemote(remotes []Remote, name, owner, repo string) bool {
	for _, remote := range remotes {
		if remote.Name != name {
			continue
		}
		if remote.Host == "github.com" && remote.Owner == owner && remote.Repo == repo {
			return true
		}
	}
	return false
}