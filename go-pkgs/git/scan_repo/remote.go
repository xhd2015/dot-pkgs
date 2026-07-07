package scan_repo

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

func ParseRemoteOwnerRepo(raw string) (owner, repo string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}

	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			return "", "", false
		}
		return ownerRepoFromPath(u.Path)
	}

	hostPart := raw
	if at := strings.LastIndex(hostPart, "@"); at >= 0 {
		hostPart = hostPart[at+1:]
	}
	colon := strings.Index(hostPart, ":")
	if colon < 0 {
		return "", "", false
	}
	return ownerRepoFromPath(hostPart[colon+1:])
}

func ownerRepoFromPath(path string) (owner, repo string, ok bool) {
	path = strings.Trim(path, "/")
	if path == "" {
		return "", "", false
	}
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", false
	}
	owner = parts[len(parts)-2]
	repo = strings.TrimSuffix(parts[len(parts)-1], ".git")
	if owner == "" || repo == "" {
		return "", "", false
	}
	return owner, repo, true
}

func remoteHost(raw string) string {
	if raw == "" {
		return ""
	}
	if u, err := url.Parse(raw); err == nil && u.Host != "" {
		return strings.ToLower(u.Hostname())
	}
	if strings.Contains(raw, "://") {
		return ""
	}

	hostPart := raw
	if at := strings.LastIndex(hostPart, "@"); at >= 0 {
		hostPart = hostPart[at+1:]
	}
	if colon := strings.Index(hostPart, ":"); colon >= 0 {
		hostPart = hostPart[:colon]
	} else if slash := strings.Index(hostPart, "/"); slash >= 0 {
		return ""
	}
	return strings.ToLower(strings.Trim(hostPart, "[]"))
}

func listRemotes(ctx context.Context, repoPath string) ([]Remote, error) {
	namesOut, err := gitOutput(ctx, repoPath, "remote")
	if err != nil {
		return nil, err
	}
	if namesOut == "" {
		return nil, nil
	}

	var remotes []Remote
	for _, name := range strings.Split(namesOut, "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		remoteURL, found, err := gitOptionalOutput(ctx, repoPath, "config", "--get", "remote."+name+".url")
		if err != nil {
			return nil, err
		}
		if !found || remoteURL == "" {
			continue
		}
		owner, repo, _ := ParseRemoteOwnerRepo(remoteURL)
		remotes = append(remotes, Remote{
			Name:  name,
			URL:   remoteURL,
			Host:  remoteHost(remoteURL),
			Owner: owner,
			Repo:  repo,
		})
	}
	return remotes, nil
}

func listWorktrees(ctx context.Context, repoPath string) ([]Worktree, error) {
	entries, err := worktree.ListCtx(ctx, repoPath)
	if err != nil {
		return nil, err
	}
	return scanWorktreesFromEntries(entries, repoPath), nil
}

func scanWorktreesFromEntries(entries []worktree.Entry, mainPath string) []Worktree {
	mainClean := absLikePath(mainPath)
	worktrees := make([]Worktree, 0, len(entries))
	for _, entry := range entries {
		path := absLikePath(entry.Path)
		head := entry.HEAD
		if head == "" && entry.Branch != "" {
			head = "refs/heads/" + entry.Branch
		}
		worktrees = append(worktrees, Worktree{
			Path:   path,
			Head:   head,
			IsMain: path == mainClean,
		})
	}
	return worktrees
}

func absLikePath(path string) string {
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "/private/var/") {
		alt := "/var/" + path[len("/private/var/"):]
		if _, err := os.Stat(alt); err == nil {
			path = alt
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return filepath.Clean(abs)
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	return cmd.Run(ctx, dir, args...)
}

func gitOptionalOutput(ctx context.Context, dir string, args ...string) (string, bool, error) {
	return cmd.RunOptional(ctx, dir, args...)
}