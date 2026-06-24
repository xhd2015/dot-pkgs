package scan_repo

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	out, err := gitOutput(ctx, repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktrees(out, repoPath), nil
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

func parseWorktrees(output, mainPath string) []Worktree {
	mainClean := absLikePath(mainPath)
	var worktrees []Worktree
	var current *Worktree

	flush := func() {
		if current == nil {
			return
		}
		current.Path = absLikePath(current.Path)
		current.IsMain = current.Path == mainClean
		worktrees = append(worktrees, *current)
		current = nil
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			flush()
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			current = &Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case current != nil && strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case current != nil && strings.HasPrefix(line, "branch "):
			current.Head = strings.TrimPrefix(line, "branch ")
		}
	}
	flush()
	return worktrees
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	out, ok, err := gitOptionalOutput(ctx, dir, args...)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("git %s in %s returned no output", strings.Join(args, " "), dir)
	}
	return out, nil
}

func gitOptionalOutput(ctx context.Context, dir string, args ...string) (string, bool, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 && text == "" {
			return "", false, nil
		}
		return "", false, fmt.Errorf("git %s in %s: %w\n%s", strings.Join(args, " "), dir, err, output)
	}
	return text, true, nil
}