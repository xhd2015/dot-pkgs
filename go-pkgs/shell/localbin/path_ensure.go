package localbin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	checkerBegin = "# ----- BEGIN ~/.local/bin checker -----"
	checkerEnd   = "# ----- END ~/.local/bin checker -----"
)

// CheckerBlock is the canonical PATH marker written into shell rc files.
func CheckerBlock() string {
	return strings.Join([]string{
		checkerBegin,
		`case ":$PATH:" in`,
		`  *":$HOME/.local/bin:"*) ;;`,
		`  *) export PATH="$HOME/.local/bin:$PATH" ;;`,
		`esac`,
		checkerEnd,
	}, "\n") + "\n"
}

// EnsureOpts configures EnsureOnPATH.
type EnsureOpts struct {
	// Home is the user home used to resolve ~/.local/bin and rc paths.
	// Empty → os.UserHomeDir.
	Home string
	// DestDir is the install destination. Ensure runs only when it equals
	// DefaultDir(Home). Empty → DefaultDir(Home) (always runs).
	DestDir string
	// DryRun probes rc files and prints [dry-run] would/skip lines; writes nothing.
	DryRun bool
	// Stdout receives dry-run plan lines. nil → os.Stdout when DryRun.
	Stdout io.Writer
	// Stderr receives live status and warning lines. nil → os.Stderr.
	Stderr io.Writer
}

// EnsureResult summarizes EnsureOnPATH.
type EnsureResult struct {
	// Updated are rc file paths that were created, appended, or replaced
	// (live), or that would be mutated (dry-run).
	Updated []string
	// Skipped is true when DestDir is not the default ~/.local/bin.
	Skipped bool
}

func (o EnsureOpts) stderr() io.Writer {
	if o.Stderr != nil {
		return o.Stderr
	}
	return os.Stderr
}

func (o EnsureOpts) stdout() io.Writer {
	if o.Stdout != nil {
		return o.Stdout
	}
	return os.Stdout
}

func (o EnsureOpts) home() (string, error) {
	home := strings.TrimSpace(o.Home)
	if home != "" {
		return home, nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	home = strings.TrimSpace(h)
	if home == "" {
		return "", fmt.Errorf("HOME is unset")
	}
	return home, nil
}

// EnsureOnPATH updates bash/zsh rc files so ~/.local/bin is on PATH when
// DestDir is the default local bin. Non-fatal: writes warnings to Stderr and
// never returns an error for individual rc update failures.
// When DryRun is set, probes the same files and prints [dry-run] would/skip
// lines to Stdout without writing.
func EnsureOnPATH(opts EnsureOpts) EnsureResult {
	home, err := opts.home()
	if err != nil {
		return EnsureResult{Skipped: true}
	}
	dest := strings.TrimSpace(opts.DestDir)
	if dest == "" {
		d, err := DefaultDir(home)
		if err != nil {
			return EnsureResult{Skipped: true}
		}
		dest = d
	}
	if !IsDefaultDest(dest, home) {
		return EnsureResult{Skipped: true}
	}

	errW := opts.stderr()
	outW := opts.stdout()
	var updated []string
	createFiles := []string{
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
	}
	existOnly := []string{
		filepath.Join(home, ".zprofile"),
		filepath.Join(home, ".profile"),
	}
	for _, f := range createFiles {
		if opts.DryRun {
			action, err := planCheckerInFile(f, true)
			if err != nil {
				fmt.Fprintf(errW, "[dry-run] warning: could not probe %s: %v\n", f, err)
				continue
			}
			printDryRunRC(outW, f, action)
			if actionMutates(action) {
				updated = append(updated, f)
			}
			continue
		}
		action, err := ensureCheckerInFile(f, true, errW)
		if err != nil {
			fmt.Fprintf(errW, "warning: could not update %s; add this to your shell rc:\n  export PATH=\"$HOME/.local/bin:$PATH\"\n", f)
			continue
		}
		if actionMutates(action) {
			updated = append(updated, f)
		}
	}
	for _, f := range existOnly {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		if opts.DryRun {
			action, err := planCheckerInFile(f, false)
			if err != nil {
				fmt.Fprintf(errW, "[dry-run] warning: could not probe %s: %v\n", f, err)
				continue
			}
			printDryRunRC(outW, f, action)
			if actionMutates(action) {
				updated = append(updated, f)
			}
			continue
		}
		action, err := ensureCheckerInFile(f, false, errW)
		if err != nil {
			fmt.Fprintf(errW, "warning: could not update %s; add this to your shell rc:\n  export PATH=\"$HOME/.local/bin:$PATH\"\n", f)
			continue
		}
		if actionMutates(action) {
			updated = append(updated, f)
		}
	}

	if opts.DryRun {
		return EnsureResult{Updated: updated}
	}
	if len(updated) > 0 {
		names := make([]string, len(updated))
		for i, f := range updated {
			names[i] = "~/" + filepath.Base(f)
		}
		fmt.Fprintf(errW, "Added ~/.local/bin to PATH in %s\n", strings.Join(names, ", "))
		fmt.Fprintf(errW, "Open a new terminal, or run: export PATH=\"$HOME/.local/bin:$PATH\"\n")
		return EnsureResult{Updated: updated}
	}
	fmt.Fprintf(errW, "PATH already includes ~/.local/bin\n")
	return EnsureResult{}
}

func actionMutates(action string) bool {
	switch action {
	case "created", "appended", "replaced":
		return true
	default:
		return false
	}
}

func printDryRunRC(w io.Writer, path, action string) {
	name := "~/" + filepath.Base(path)
	switch action {
	case "created":
		fmt.Fprintf(w, "[dry-run] would create %s (PATH checker)\n", name)
	case "appended":
		fmt.Fprintf(w, "[dry-run] would append %s (PATH checker)\n", name)
	case "replaced":
		fmt.Fprintf(w, "[dry-run] would replace %s (PATH checker)\n", name)
	case "unchanged":
		fmt.Fprintf(w, "[dry-run] skip: %s (PATH checker already present)\n", name)
	case "skipped_missing":
		// exist-only missing: live ignores; dry-run stays quiet
	}
}

// planCheckerInFile returns the action EnsureCheckerInFile would take without writing.
func planCheckerInFile(path string, create bool) (action string, err error) {
	canonical := CheckerBlock()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if !create {
				return "skipped_missing", nil
			}
			return "created", nil
		}
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "skipped_missing", nil
	}
	_, action, _ = ApplyChecker(string(data), canonical)
	return action, nil
}

func splitRCLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if s == "" {
		return nil
	}
	s = strings.TrimSuffix(s, "\n")
	return strings.Split(s, "\n")
}

func joinRCLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func linesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type checkerBlock struct {
	start int
	end   int
}

// ApplyChecker returns rewritten file content and action:
// unchanged | appended | replaced.
func ApplyChecker(content, canonical string) (newContent, action string, warnOrphan bool) {
	lines := splitRCLines(content)
	canonLines := splitRCLines(canonical)

	var blocks []checkerBlock
	var orphans []int
	open := -1
	for i, line := range lines {
		switch {
		case line == checkerBegin:
			if open >= 0 {
				orphans = append(orphans, open)
			}
			open = i
		case line == checkerEnd && open >= 0:
			blocks = append(blocks, checkerBlock{start: open, end: i})
			open = -1
		}
	}
	if open >= 0 {
		orphans = append(orphans, open)
	}
	warnOrphan = len(orphans) > 0

	if len(blocks) == 1 && !warnOrphan {
		got := lines[blocks[0].start : blocks[0].end+1]
		if linesEqual(got, canonLines) {
			return content, "unchanged", false
		}
	}

	if len(blocks) == 0 {
		out := append([]string{}, lines...)
		if len(out) > 0 && out[len(out)-1] != "" {
			out = append(out, "")
		}
		out = append(out, canonLines...)
		return joinRCLines(out), "appended", warnOrphan
	}

	out := make([]string, 0, len(lines))
	emitted := false
	bi := 0
	i := 0
	for i < len(lines) {
		if bi < len(blocks) && i == blocks[bi].start {
			if !emitted {
				out = append(out, canonLines...)
				emitted = true
			}
			i = blocks[bi].end + 1
			bi++
			continue
		}
		out = append(out, lines[i])
		i++
	}
	return joinRCLines(out), "replaced", warnOrphan
}

// EnsureCheckerInFile creates or updates path with the canonical checker block.
// create=false skips missing files (skipped_missing). Exported for tests and
// callers that patch a single file.
func EnsureCheckerInFile(path string, create bool) (action string, err error) {
	return ensureCheckerInFile(path, create, io.Discard)
}

func ensureCheckerInFile(path string, create bool, warnOut io.Writer) (action string, err error) {
	if warnOut == nil {
		warnOut = io.Discard
	}
	canonical := CheckerBlock()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if !create {
				return "skipped_missing", nil
			}
			if err := os.WriteFile(path, []byte(canonical), 0o644); err != nil {
				return "", err
			}
			return "created", nil
		}
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "skipped_missing", nil
	}

	newContent, action, warnOrphan := ApplyChecker(string(data), canonical)
	if warnOrphan {
		if action == "appended" {
			fmt.Fprintf(warnOut, "warning: %s has a BEGIN ~/.local/bin checker without END; appended a new block\n", path)
		} else {
			fmt.Fprintf(warnOut, "warning: %s has a BEGIN ~/.local/bin checker without END; left it in place\n", path)
		}
	}
	if action == "unchanged" {
		return action, nil
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(newContent); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return "", err
	}
	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return "", err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return "", err
	}
	return action, nil
}
