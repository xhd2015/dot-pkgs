package libmark

import (
	"context"
	"strings"

	gitcmd "github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
)

// CaptureGit snapshots git identity for dir. Returns nil when dir is not a repo.
func CaptureGit(dir string) *GitInfo {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil
	}
	ctx := context.Background()
	toplevel, err := gitcmd.Run(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil || toplevel == "" {
		return nil
	}
	gitdir, err := gitcmd.Run(ctx, dir, "rev-parse", "--absolute-git-dir")
	if err != nil || gitdir == "" {
		return nil
	}
	g := &GitInfo{
		Toplevel: toplevel,
		GitDir:   gitdir,
	}
	if commit, err := gitcmd.Run(ctx, dir, "rev-parse", "HEAD"); err == nil {
		g.Commit = commit
	}
	if branch, err := gitcmd.Run(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		g.Branch = branch
	}
	if origin, ok, err := gitcmd.RunOptional(ctx, dir, "config", "--get", "remote.origin.url"); err == nil && ok {
		g.Origin = origin
	}
	return g
}
