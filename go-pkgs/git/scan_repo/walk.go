package scan_repo

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/remotefs"
)

func walkRoot(ctx context.Context, root string, maxDepth int, ignore ignoreConfig, verbose bool, stderr io.Writer, onRepo func(Repo) error) ([]Repo, error) {
	if remote, err := remotefs.IsRemoteBackedPath(root); err != nil {
		maybeWarnSkip(verbose, stderr, root, err)
		return nil, nil
	} else if remote {
		maybeWarnSkipRemote(verbose, stderr, root)
		return []Repo{}, nil
	}

	var repos []Repo
	var repoRoots []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if isPermissionError(walkErr) {
				maybeWarnSkip(verbose, stderr, path, walkErr)
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
			if shouldSkipDirBasename(path, d.Name(), repoRoots, ignore) {
				return filepath.SkipDir
			}
			if maxDepth > 0 && depthFromRoot(root, path) > maxDepth {
				return filepath.SkipDir
			}
			if remote, err := remotefs.IsRemoteBackedPath(path); err != nil {
				maybeWarnSkip(verbose, stderr, path, err)
				return filepath.SkipDir
			} else if remote {
				maybeWarnSkipRemote(verbose, stderr, path)
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
			repo := Repo{
				Path:  path,
				Name:  filepath.Base(path),
				Error: resolveErr.Error(),
			}
			if onRepo != nil {
				if err := onRepo(repo); err != nil {
					return err
				}
			} else {
				repos = append(repos, repo)
			}
			return nil
		}

		repo := Repo{
			Path:     path,
			Name:     filepath.Base(path),
			GitDir:   gitDir,
			RepoType: repoType,
		}
		if onRepo != nil {
			if err := onRepo(repo); err != nil {
				return err
			}
		} else {
			repos = append(repos, repo)
		}
		repoRoots = append(repoRoots, path)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func isPermissionError(err error) bool {
	return os.IsPermission(err)
}

func maybeWarnSkip(verbose bool, stderr io.Writer, path string, err error) {
	if !verbose {
		return
	}
	fmt.Fprintf(stderrWriter(stderr), "\nwarning: skipping\n%s: %v", path, err)
}

func maybeWarnSkipRemote(verbose bool, stderr io.Writer, path string) {
	if !verbose {
		return
	}
	fmt.Fprintf(stderrWriter(stderr), "\nwarning: skipping remote-backed filesystem\n%s", path)
}

func stderrWriter(stderr io.Writer) io.Writer {
	if stderr != nil {
		return stderr
	}
	return os.Stderr
}

func shouldSkipDirBasename(path, name string, repoRoots []string, ignore ignoreConfig) bool {
	if isInsideRepoCheckout(path, repoRoots) {
		return name == ".git"
	}
	_, skip := ignore.basenames[name]
	return skip
}

func isInsideRepoCheckout(path string, repoRoots []string) bool {
	clean := filepath.Clean(path)
	sep := string(filepath.Separator)
	for _, root := range repoRoots {
		root = filepath.Clean(root)
		if clean == root {
			return true
		}
		rel, err := filepath.Rel(root, clean)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+sep) {
			continue
		}
		gitDir := filepath.Join(root, ".git")
		if clean == gitDir || strings.HasPrefix(clean, gitDir+sep) {
			continue
		}
		return true
	}
	return false
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