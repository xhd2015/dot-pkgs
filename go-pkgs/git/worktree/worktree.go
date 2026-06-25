package worktree

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Entry represents one row from `git worktree list --porcelain`.
type Entry struct {
	Path   string
	Branch string // empty when detached
	HEAD   string
	IsMain bool   // true when .git is a directory (main checkout)
}

// IsDead reports whether the worktree directory no longer exists on disk.
func IsDead(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

// samePath reports whether two paths refer to the same location, resolving
// symlinks in parent directories when the target itself may not exist.
func samePath(a, b string) bool {
	a = normalizePath(a)
	b = normalizePath(b)
	return a == b
}

func normalizePath(path string) string {
	path = filepath.Clean(path)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	if resolvedDir, err := filepath.EvalSymlinks(dir); err == nil {
		return filepath.Join(resolvedDir, base)
	}
	return path
}

// IsLinked reports whether path is a linked worktree (.git is a file, not a directory).
func IsLinked(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.Mode().IsRegular()
}

// IsMainRepo reports whether path is a main git checkout (.git is a directory).
func IsMainRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}

// ReadMainRepo resolves the main repository path from a linked worktree's .git file.
func ReadMainRepo(linkedPath string) (string, error) {
	gitFile := filepath.Join(linkedPath, ".git")
	content, err := os.ReadFile(gitFile)
	if err != nil {
		return "", fmt.Errorf("read .git file: %w", err)
	}
	s := strings.TrimSpace(string(content))
	const prefix = "gitdir: "
	if !strings.HasPrefix(s, prefix) {
		return "", fmt.Errorf("unexpected .git file format in worktree %s", linkedPath)
	}
	gitdir := strings.TrimSpace(s[len(prefix):])
	// gitdir is <mainRepo>/.git/worktrees/<name>
	mainRepo := filepath.Dir(filepath.Dir(filepath.Dir(gitdir)))
	return mainRepo, nil
}

// ReadBranch returns the current branch name, or "HEAD" when detached.
func ReadBranch(worktreePath string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --abbrev-ref HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ResolveMainRepo returns the main repository for path, whether path is a main
// checkout or a linked worktree.
func ResolveMainRepo(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if IsLinked(abs) {
		return ReadMainRepo(abs)
	}
	if IsMainRepo(abs) {
		return abs, nil
	}
	return "", fmt.Errorf("%s is not a git repository", abs)
}

// List returns all worktrees including the main checkout.
func List(repoPath string) ([]Entry, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	entries := parsePorcelain(string(out))
	for i := range entries {
		entries[i].IsMain = IsMainRepo(entries[i].Path)
	}
	return entries, nil
}

// ListLinked returns linked worktrees only (excludes the main checkout).
func ListLinked(repoPath string) ([]Entry, error) {
	all, err := List(repoPath)
	if err != nil {
		return nil, err
	}
	var linked []Entry
	for _, e := range all {
		if !e.IsMain {
			linked = append(linked, e)
		}
	}
	return linked, nil
}

func parsePorcelain(output string) []Entry {
	var entries []Entry
	scanner := bufio.NewScanner(strings.NewReader(output))
	var current Entry
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			if current.Path != "" {
				entries = append(entries, current)
			}
			current = Entry{Path: line[len("worktree "):]}
		case strings.HasPrefix(line, "HEAD "):
			current.HEAD = line[len("HEAD "):]
		case strings.HasPrefix(line, "branch "):
			ref := line[len("branch "):]
			const prefix = "refs/heads/"
			if strings.HasPrefix(ref, prefix) {
				current.Branch = ref[len(prefix):]
			}
		}
	}
	if current.Path != "" {
		entries = append(entries, current)
	}
	return entries
}