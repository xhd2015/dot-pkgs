package scan_repo

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func walkRoot(ctx context.Context, root string, maxDepth int, ignore ignoreConfig, verbose bool) ([]Repo, error) {
	var repos []Repo

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

		repos = append(repos, Repo{
			Path:     path,
			Name:     filepath.Base(path),
			GitDir:   gitDir,
			RepoType: repoType,
		})

		return filepath.SkipDir
	})
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func isPermissionError(err error) bool {
	return os.IsPermission(err)
}

func maybeWarnSkip(verbose bool, path string, err error) {
	if !verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "\nwarning: skipping\n%s: %v", path, err)
}

func depthFromRoot(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator)) + 1
}

func resolveGitDir(checkoutPath, gitPath string, info fs.FileInfo) (string, RepoType, error) {
	if info.IsDir() {
		return filepath.Clean(gitPath), RepoTypeMain, nil
	}

	data, err := os.ReadFile(gitPath)
	if err != nil {
		return "", "", err
	}

	gitdir, ok := parseGitLink(string(data))
	if !ok {
		return "", "", nil
	}

	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(checkoutPath, gitdir)
	}
	gitdir = filepath.Clean(gitdir)

	if idx := strings.Index(gitdir, ".git"+string(filepath.Separator)+"worktrees"+string(filepath.Separator)); idx >= 0 {
		gitdir = gitdir[:idx+len(".git")]
	}

	return gitdir, RepoTypeWorktree, nil
}

func parseGitLink(content string) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gitdir:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "gitdir:")), true
		}
	}
	return "", false
}