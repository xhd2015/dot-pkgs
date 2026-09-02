package rg

import (
	"context"
	"fmt"
	"io"
)

// EnsureOpts configures Ensure.
type EnsureOpts struct {
	Discover DiscoverOpts
	Install  InstallOpts
	// Stderr receives human notices (install progress). Nil → discard.
	Stderr io.Writer
	// Notice writes a notice line body (without "notice: " prefix). Nil →
	// fmt.Fprintln(Stderr, "notice: "+body) when Stderr set.
	Notice func(body string)
}

// EnsureResult is the outcome of Ensure.
type EnsureResult struct {
	Action   string // "noop" | "install"
	Selected CLI
	All      []CLI // all discovered before/after (post-ensure set)
	Install  InstallResult
}

// Ensure resolves the newest rg, installing from GitHub when none is found.
// It does not auto-upgrade an existing older install.
func Ensure(ctx context.Context, opts EnsureOpts) (EnsureResult, error) {
	var result EnsureResult
	if ctx == nil {
		ctx = context.Background()
	}
	notice := opts.Notice
	if notice == nil && opts.Stderr != nil {
		notice = func(body string) {
			fmt.Fprintln(opts.Stderr, "notice: "+body)
		}
	}
	if notice == nil {
		notice = func(string) {}
	}

	found, err := Found(ctx, opts.Discover)
	if err != nil {
		return result, err
	}
	if len(found) > 0 {
		selected, err := Newest(ctx, opts.Discover)
		if err != nil {
			// fall back to first found
			selected = found[0]
		}
		result.Action = "noop"
		result.Selected = selected
		result.All = found
		notice(FormatUsingNotice(selected, found))
		return result, nil
	}

	notice("rg not found; installing BurntSushi/ripgrep from GitHub releases…")
	inst, err := InstallLatest(ctx, opts.Install)
	if err != nil {
		result.Action = "install"
		return result, err
	}
	result.Action = "install"
	result.Install = inst
	notice(fmt.Sprintf("installed rg %s → %s", inst.Version, inst.BinPath))

	// Re-discover so well-known/login paths pick up the new binary.
	found, err = Found(ctx, opts.Discover)
	if err != nil {
		return result, err
	}
	if len(found) == 0 {
		// Directly use installed path if discovery still empty (PATH not updated).
		selected := CLI{Path: inst.BinPath, Version: inst.Version, Via: viaWellKnown}
		result.Selected = selected
		result.All = []CLI{selected}
		notice(FormatUsingNotice(selected, nil))
		return result, nil
	}
	selected, err := Newest(ctx, opts.Discover)
	if err != nil {
		selected = found[0]
	}
	result.Selected = selected
	result.All = found
	notice(FormatUsingNotice(selected, found))
	return result, nil
}
