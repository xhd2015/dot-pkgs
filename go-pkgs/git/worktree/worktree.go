package worktree

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
	gitops "github.com/xhd2015/gitops/git"
	gitopsWorktree "github.com/xhd2015/gitops/git/worktree"
)

// GitVerboseLogger is invoked before major git subprocesses when wrk -v is set.
// args are the full git argv including -C <dir> when applicable.
var GitVerboseLogger func(args []string)

func logGitVerbose(args []string) {
	if GitVerboseLogger != nil {
		GitVerboseLogger(args)
	}
}

// Entry represents one row from `git worktree list --porcelain`.
// Public go-pkgs type; values are converted from gitops worktree.Entry.
type Entry struct {
	Path   string
	Branch string // empty when detached
	HEAD   string
	IsMain bool // true when .git is a directory (main checkout)
}

func entryFromGitops(e gitopsWorktree.Entry) Entry {
	return Entry{
		Path:   e.Path,
		Branch: e.Branch,
		HEAD:   e.HEAD,
		IsMain: e.IsMain,
	}
}

func entriesFromGitops(es []gitopsWorktree.Entry) []Entry {
	if es == nil {
		return nil
	}
	out := make([]Entry, len(es))
	for i, e := range es {
		out[i] = entryFromGitops(e)
	}
	return out
}

// IsDead reports whether the worktree directory no longer exists on disk.
func IsDead(path string) bool {
	return gitopsWorktree.IsDead(path)
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

// IsInsideWorkTree reports whether path is inside a git work tree.
func IsInsideWorkTree(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// ShowToplevel returns the root of the git work tree containing path.
func ShowToplevel(path string) (string, error) {
	out, err := gitops.ShowToplevel(path)
	if err != nil {
		return "", err
	}
	top := strings.TrimSpace(out)
	abs, err := filepath.Abs(top)
	if err != nil {
		return "", fmt.Errorf("resolve work tree root: %w", err)
	}
	return abs, nil
}

// IsLinked reports whether path is a linked worktree (.git is a file, not a directory).
func IsLinked(path string) bool {
	return gitopsWorktree.IsLinked(path)
}

// IsMainRepo reports whether path is a main git checkout (.git is a directory).
func IsMainRepo(path string) bool {
	return gitopsWorktree.IsMainRepo(path)
}

// ReadMainRepo resolves the main repository path from a linked worktree's .git file.
func ReadMainRepo(linkedPath string) (string, error) {
	return gitopsWorktree.ReadMainRepo(linkedPath)
}

// ReadBranch returns the current branch name, or "HEAD" when detached.
func ReadBranch(worktreePath string) (string, error) {
	return ReadBranchCtx(context.Background(), worktreePath)
}

// ReadBranchCtx is ReadBranch with cancellation support.
func ReadBranchCtx(ctx context.Context, worktreePath string) (string, error) {
	branch, err := cmd.Run(ctx, worktreePath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git rev-parse --abbrev-ref HEAD: %w", err)
	}
	return branch, nil
}

// ResolveMainRepo returns the main repository for path, whether path is a main
// checkout or a linked worktree.
func ResolveMainRepo(path string) (string, error) {
	return gitopsWorktree.ResolveMainRepo(path)
}

// List returns all worktrees including the main checkout.
func List(repoPath string) ([]Entry, error) {
	return ListCtx(context.Background(), repoPath)
}

// ListCtx is List with cancellation support.
// Low-level inventory is delegated to gitops; ctx is checked before the call
// (gitops List itself has no context).
func ListCtx(ctx context.Context, repoPath string) ([]Entry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := gitopsWorktree.List(repoPath)
	if err != nil {
		return nil, err
	}
	return entriesFromGitops(entries), nil
}

// ListLinked returns linked worktrees only (excludes the main checkout).
func ListLinked(repoPath string) ([]Entry, error) {
	entries, err := gitopsWorktree.ListLinked(repoPath)
	if err != nil {
		return nil, err
	}
	return entriesFromGitops(entries), nil
}

// WorktreesOnBranch returns registered worktrees whose Branch equals branch.
// Detached entries (Branch empty) never match. Multiple matches are returned
// as data only — no policy error.
func WorktreesOnBranch(repoPath, branch string) ([]Entry, error) {
	entries, err := gitopsWorktree.WorktreesOnBranch(repoPath, branch)
	if err != nil {
		return nil, err
	}
	return entriesFromGitops(entries), nil
}

// BranchSharedError is returned when a named branch is checked out in more than
// one registered worktree (including dead/prunable entries). Callers may wrap
// with product-specific refuse wording; Error() is generic library text.
type BranchSharedError struct {
	Branch   string
	MainRepo string
	Entries  []Entry
}

func (e *BranchSharedError) Error() string {
	if e == nil {
		return "branch is checked out in multiple worktrees"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "branch %q is checked out in multiple worktrees:", e.Branch)
	for _, ent := range e.Entries {
		if IsDead(ent.Path) {
			fmt.Fprintf(&b, "\n  %s (missing; prune with: git -C %s worktree prune)", ent.Path, e.MainRepo)
		} else {
			fmt.Fprintf(&b, "\n  %s", ent.Path)
		}
	}
	return b.String()
}

// EnsureBranchExclusive returns *BranchSharedError when branch has more than one
// registered worktree checkout under mainRepo. Detached HEAD (empty branch or
// "HEAD") is skipped. Dead registrations count toward the multi-checkout total.
func EnsureBranchExclusive(mainRepo, branch string) error {
	if branch == "" || branch == "HEAD" {
		return nil
	}
	entries, err := WorktreesOnBranch(mainRepo, branch)
	if err != nil {
		return err
	}
	if len(entries) > 1 {
		return &BranchSharedError{
			Branch:   branch,
			MainRepo: mainRepo,
			Entries:  entries,
		}
	}
	return nil
}

// ParseListPorcelain parses `git worktree list --porcelain` output into entries.
func ParseListPorcelain(output string) []Entry {
	return entriesFromGitops(gitopsWorktree.ParseListPorcelain(output))
}
