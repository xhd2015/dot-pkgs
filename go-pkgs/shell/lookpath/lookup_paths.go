package lookpath

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LookupItem is one name's resolution outcome from LookupPaths.
type LookupItem struct {
	Name    string // requested name
	Path    string // absolute binary path if found; "" if Missing
	Missing bool   // true if not found
	From    string // "bash" | "zsh" if found via login shell; "" otherwise
}

// LookupItems is an ordered list of LookupItem (one per input name).
type LookupItems []LookupItem

// LookupPaths batch-resolves bare CLI names with the same cheap stages as Look
// (PATH → ExtraDirs → DefaultDirs → ExtraCandidates), then login for remaining.
//
// One item per input name, same order; no name dedupe. Empty names → empty
// items and nil error. An empty-string element returns an error. Missing names
// are best-effort (Missing=true), never a LookupPaths error.
//
// Login From is the shell basename only ("bash"/"zsh"), not "login_shell:…".
func LookupPaths(names []string, opts Options) (LookupItems, error) {
	if len(names) == 0 {
		return LookupItems{}, nil
	}
	for _, name := range names {
		if name == "" {
			return nil, fmt.Errorf("lookpath: empty name")
		}
	}

	isExec := opts.IsExecutable
	if isExec == nil {
		isExec = IsExecutable
	}
	lookPath := opts.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	items := make(LookupItems, len(names))
	var remaining []int // indices still unresolved after cheap stages

	for i, name := range names {
		items[i] = LookupItem{Name: name}
		if path, ok := resolveCheap(name, opts, lookPath, isExec); ok {
			items[i].Path = path
			continue
		}
		remaining = append(remaining, i)
	}

	if len(remaining) > 0 {
		resolveLogin(items, remaining, opts)
	}

	for i := range items {
		if items[i].Path == "" {
			items[i].Missing = true
			items[i].From = ""
		}
	}
	return items, nil
}

// Dirs returns unique cleaned parent directories of found paths, first-seen order.
func (items LookupItems) Dirs() []string {
	seen := make(map[string]struct{}, len(items))
	dirs := make([]string, 0, len(items))
	for _, it := range items {
		if it.Missing || it.Path == "" {
			continue
		}
		d := filepath.Clean(filepath.Dir(it.Path))
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		dirs = append(dirs, d)
	}
	return dirs
}

// DirsEnv joins Dirs() with os.PathListSeparator; empty string when no dirs.
func (items LookupItems) DirsEnv() string {
	dirs := items.Dirs()
	if len(dirs) == 0 {
		return ""
	}
	return strings.Join(dirs, string(os.PathListSeparator))
}

// resolveCheap runs PATH → ExtraDirs → DefaultDirs → ExtraCandidates.
func resolveCheap(name string, opts Options, lookPath func(string) (string, error), isExec func(string) bool) (string, bool) {
	if p, err := lookPath(name); err == nil && p != "" {
		return p, true
	}
	for _, dir := range opts.ExtraDirs {
		if dir == "" {
			continue
		}
		p := filepath.Join(dir, name)
		if isExec(p) {
			return p, true
		}
	}
	for _, dir := range DefaultDirs(opts.Home) {
		p := filepath.Join(dir, name)
		if isExec(p) {
			return p, true
		}
	}
	for _, cand := range opts.ExtraCandidates {
		if cand == "" {
			continue
		}
		if isExec(cand) {
			return cand, true
		}
	}
	return "", false
}

// resolveLogin fills still-missing items via login shells.
// Prefers one batch script per shell for multiple remaining names; falls back
// to per-name "command -v" probes. From is set to shell basename only.
func resolveLogin(items LookupItems, remaining []int, opts Options) {
	shells := opts.Shells
	if len(shells) == 0 {
		shells = []string{"bash", "zsh"}
	}
	runLogin := opts.RunLogin
	if runLogin == nil {
		timeout := opts.Timeout
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		runLogin = defaultRunLogin(timeout)
	}
	env := minimalLoginEnv(opts.Home)

	for _, shell := range shells {
		if shell == "" {
			continue
		}
		need := stillMissing(items, remaining)
		if len(need) == 0 {
			return
		}
		from := filepath.Base(shell)

		// Prefer batch one shell script when multiple names remain.
		if len(need) > 1 {
			cmd := batchCommandV(need, items)
			if stdout, err := runLogin(shell, cmd, env); err == nil {
				if lines := splitExactLines(stdout, len(need)); lines != nil {
					for j, idx := range need {
						p := strings.TrimSpace(lines[j])
						if p == "" {
							continue
						}
						items[idx].Path = p
						items[idx].From = from
					}
					// Batch parse accepted; unresolved names try next shell.
					continue
				}
			}
		}

		// Per-name probes (single remaining, or batch parse unavailable).
		for _, idx := range stillMissing(items, need) {
			stdout, err := runLogin(shell, "command -v "+items[idx].Name, env)
			if err != nil {
				continue
			}
			p := strings.TrimSpace(stdout)
			if p == "" {
				continue
			}
			items[idx].Path = p
			items[idx].From = from
		}
	}
}

func stillMissing(items LookupItems, indices []int) []int {
	var out []int
	for _, idx := range indices {
		if items[idx].Path == "" {
			out = append(out, idx)
		}
	}
	return out
}

// batchCommandV builds a login command that prints one line per name.
func batchCommandV(indices []int, items LookupItems) string {
	var b strings.Builder
	for i, idx := range indices {
		if i > 0 {
			b.WriteString("; ")
		}
		fmt.Fprintf(&b, `printf '%%s\n' "$(command -v %s 2>/dev/null)"`, items[idx].Name)
	}
	return b.String()
}

// splitExactLines splits stdout into exactly want lines (trailing newline OK).
// Returns nil when the line count does not match want.
func splitExactLines(stdout string, want int) []string {
	if want <= 0 {
		return nil
	}
	s := strings.TrimSuffix(stdout, "\n")
	// Empty / whitespace-only: only valid as want empty lines when want==1.
	if strings.TrimSpace(stdout) == "" {
		if want == 1 {
			return []string{""}
		}
		return nil
	}
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if len(lines) != want {
		return nil
	}
	return lines
}
