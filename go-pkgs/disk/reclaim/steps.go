package reclaim

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Tier groups steps by when they participate in a run.
//
// TierDefault steps always run. TierAggressive steps are expensive to rebuild
// or large in scope, so they only run when Enabled says so (--aggressive-all or
// an explicit opt-in flag).
type Tier int

const (
	TierDefault Tier = iota
	TierAggressive
)

// stepDef describes one cleanup step: measure, optional apply.
//
// Every step (default or aggressive) flows through the same Run loop, so a
// failing step always warns and lets the remaining steps run.
type stepDef struct {
	ID   string
	Tier Tier
	// Enabled reports whether this step participates in the run.
	// nil means "always" (all TierDefault steps).
	Enabled func(opts Options) bool
	Measure func(opts Options) int64
	Apply   func(opts Options) error
}

// allSteps returns the full registry in execution order: default steps first,
// then aggressive steps.
func allSteps() []stepDef {
	return append(defaultSteps(), aggressiveSteps()...)
}

func defaultSteps() []stepDef {
	return []stepDef{
		{
			ID: "sandbox-gocache",
			Measure: func(opts Options) int64 {
				return pathSize(filepath.Join(opts.Home, ".sandbox", "gocache"))
			},
			Apply: func(opts Options) error {
				return os.RemoveAll(filepath.Join(opts.Home, ".sandbox", "gocache"))
			},
		},
		{
			ID: "codex-logs",
			Measure: func(opts Options) int64 {
				return globSize(filepath.Join(opts.Home, ".codex", "logs_2.sqlite*"))
			},
			Apply: func(opts Options) error {
				matches, err := filepath.Glob(filepath.Join(opts.Home, ".codex", "logs_2.sqlite*"))
				if err != nil {
					return err
				}
				for _, m := range matches {
					if err := os.RemoveAll(m); err != nil && !os.IsNotExist(err) {
						return err
					}
				}
				return nil
			},
		},
		{
			ID: "home-logs",
			Measure: func(opts Options) int64 {
				return filesSize(
					filepath.Join(opts.Home, "ai-critic-server.log"),
					filepath.Join(opts.Home, "ai-critic-server-keep-alive.log"),
				)
			},
			Apply: func(opts Options) error {
				for _, name := range []string{
					"ai-critic-server.log",
					"ai-critic-server-keep-alive.log",
				} {
					p := filepath.Join(opts.Home, name)
					if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
						return err
					}
				}
				return nil
			},
		},
		{
			ID: "brew-cleanup",
			Measure: func(opts Options) int64 {
				return pathSize(filepath.Join(opts.Home, "Library", "Caches", "Homebrew"))
			},
			Apply: func(opts Options) error {
				// -s scrubs versioned cache; --prune=all drops all age-eligible
				// cache files (incl. downloads for installed formulae/casks).
				// Does not uninstall packages — only re-download cost on next use.
				if err := runExternalTool(opts, "brew", "cleanup", "-s", "--prune=all"); err != nil {
					return err
				}
				// brew leaves metadata caches behind (api/, bootsnap/, Backup/),
				// which keep the step's measured footprint non-zero while being
				// pure caches brew re-fetches on demand. Drop them so Freed
				// matches what the step promised.
				p := filepath.Join(opts.Home, "Library", "Caches", "Homebrew")
				if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
					return err
				}
				return nil
			},
		},
		{
			ID: "npm-cache",
			Measure: func(opts Options) int64 {
				return pathSize(filepath.Join(opts.Home, ".npm", "_cacache"))
			},
			Apply: func(opts Options) error {
				return runExternalTool(opts, "npm", "cache", "clean", "--force")
			},
		},
		{
			ID: "remote-agent-logs",
			Measure: func(opts Options) int64 {
				return globSize(filepath.Join(opts.Home, "Library", "Caches", "remote-agent", "*.log"))
			},
			Apply: func(opts Options) error {
				matches, err := filepath.Glob(filepath.Join(opts.Home, "Library", "Caches", "remote-agent", "*.log"))
				if err != nil {
					return err
				}
				for _, m := range matches {
					if err := os.Remove(m); err != nil && !os.IsNotExist(err) {
						return err
					}
				}
				return nil
			},
		},
		{
			ID: "podman-applehv-cache",
			Measure: func(opts Options) int64 {
				return pathSize(filepath.Join(opts.Home, ".local", "share", "containers", "podman", "machine", "applehv", "cache"))
			},
			Apply: func(opts Options) error {
				p := filepath.Join(opts.Home, ".local", "share", "containers", "podman", "machine", "applehv", "cache")
				if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
					return err
				}
				return nil
			},
		},
		{
			ID:      stepITermBakApps,
			Measure: measureITermBakApps,
			Apply:   applyITermBakApps,
		},
	}
}

// aggressiveSteps are the expensive, high-reclaim caches: build caches and
// package-manager stores whose only cost is a re-download or a rebuild.
//
// NOTE: `go clean -modcache` (~12 GiB of modules) is deliberately absent and
// must never be added: it forces network re-fetch of every module. The Go build
// cache is safe because it is pure compiled output.
func aggressiveSteps() []stepDef {
	return []stepDef{
		{
			ID:   "go-build-cache",
			Tier: TierAggressive,
			Enabled: func(opts Options) bool {
				return opts.Aggressive
			},
			Measure: func(opts Options) int64 {
				return pathSize(goBuildCachePath(opts.Home))
			},
			Apply: applyGoBuildCache,
		},
		{
			ID:   "pnpm-store",
			Tier: TierAggressive,
			Enabled: func(opts Options) bool {
				return opts.Aggressive
			},
			Measure: func(opts Options) int64 {
				return pathSize(pnpmStorePath(opts.Home))
			},
			Apply: func(opts Options) error {
				// Prunes unreferenced packages from the content-addressable
				// store; installed projects keep working.
				return runExternalTool(opts, "pnpm", "store", "prune")
			},
		},
		{
			ID:   "npm-npx",
			Tier: TierAggressive,
			Enabled: func(opts Options) bool {
				return opts.Aggressive
			},
			Measure: func(opts Options) int64 {
				return pathSize(npmNpxPath(opts.Home))
			},
			Apply: func(opts Options) error {
				// Cached `npx`/`bunx`-style one-shot packages; re-fetched on use.
				p := npmNpxPath(opts.Home)
				if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
					return err
				}
				return nil
			},
		},
		{
			ID:   "bun-cache",
			Tier: TierAggressive,
			Enabled: func(opts Options) bool {
				return opts.Aggressive
			},
			Measure: func(opts Options) int64 {
				return pathSize(bunCachePath(opts.Home))
			},
			Apply: func(opts Options) error {
				// Plain removal: `bun pm cache rm` refuses to run outside a
				// project ("No package.json was found") and the directory it
				// targets is exactly this path, so remove it directly.
				p := bunCachePath(opts.Home)
				if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
					return err
				}
				return nil
			},
		},
		{
			ID:   "codex-releases",
			Tier: TierAggressive,
			Enabled: func(opts Options) bool {
				return opts.Aggressive
			},
			Measure: func(opts Options) int64 {
				return staleCodexReleasesSize(opts.Home)
			},
			Apply: applyStaleCodexReleases,
		},
		{
			ID:   "xcache",
			Tier: TierAggressive,
			Enabled: func(opts Options) bool {
				return opts.Aggressive
			},
			Measure: func(opts Options) int64 {
				return pathSize(xcachePath(opts.Home))
			},
			Apply: func(opts Options) error {
				// {Home}/.cache is pure cache: no user data, only re-download
				// cost (puppeteer chromium, uv wheels, codex runtime bundle...).
				p := xcachePath(opts.Home)
				if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
					return err
				}
				return nil
			},
		},
		{
			ID:   stepMSPlaywright,
			Tier: TierAggressive,
			Enabled: func(opts Options) bool {
				return opts.Aggressive || opts.IncludeMSPlaywright
			},
			Measure: measureMSPlaywright,
			Apply:   applyMSPlaywright,
		},
		{
			ID:   stepOpenCodeSnapshot,
			Tier: TierAggressive,
			Enabled: func(opts Options) bool {
				return opts.Aggressive || opts.IncludeOpenCodeSnapshot
			},
			Measure: measureOpenCodeSnapshot,
			Apply:   applyOpenCodeSnapshot,
		},
	}
}

// goBuildCachePath is the default GOCACHE on macOS.
func goBuildCachePath(home string) string {
	return filepath.Join(home, "Library", "Caches", "go-build")
}

// pnpmStorePath is the default pnpm content-addressable store.
func pnpmStorePath(home string) string {
	return filepath.Join(home, "Library", "pnpm", "store")
}

// npmNpxPath holds packages fetched by `npx` one-shots.
func npmNpxPath(home string) string {
	return filepath.Join(home, ".npm", "_npx")
}

// bunCachePath is bun's install cache ("bun pm cache rm" target).
func bunCachePath(home string) string {
	return filepath.Join(home, ".bun", "install", "cache")
}

// xcachePath is the XDG-style {Home}/.cache.
func xcachePath(home string) string {
	return filepath.Join(home, ".cache")
}

// codexReleasesDir holds unpacked codex CLI runtime bundles.
func codexReleasesDir(home string) string {
	return filepath.Join(home, ".codex", "packages", "standalone", "releases")
}

// applyGoBuildCache runs `go clean -cache` (respects a custom GOCACHE).
// Without a go toolchain the cache directory is still safe to drop directly.
func applyGoBuildCache(opts Options) error {
	if _, err := lookTool(opts, "go"); err != nil {
		p := goBuildCachePath(opts.Home)
		if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return runExternalTool(opts, "go", "clean", "-cache")
}

// staleCodexReleases lists codex release dirs that are not the newest, sorted
// oldest-first. The newest release is always kept (it is the running CLI), and
// so is a release dir containing the currently executing binary.
func staleCodexReleases(home string) []string {
	dir := codexReleasesDir(home)
	dirents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	versions := make([]string, 0, len(dirents))
	for _, e := range dirents {
		if !e.IsDir() {
			continue
		}
		if _, ok := codexReleaseVersion(e.Name()); !ok {
			continue
		}
		versions = append(versions, e.Name())
	}
	if len(versions) < 2 {
		return nil
	}
	sort.Slice(versions, func(i, j int) bool {
		return compareCodexVersions(versions[i], versions[j]) < 0
	})
	newest := versions[len(versions)-1]
	self, _ := os.Executable()

	stale := make([]string, 0, len(versions)-1)
	for _, name := range versions[:len(versions)-1] {
		p := filepath.Join(dir, name)
		if name == newest {
			continue
		}
		if self != "" && strings.HasPrefix(self, p+string(os.PathSeparator)) {
			continue // never delete the release we are running from
		}
		stale = append(stale, p)
	}
	return stale
}

// staleCodexReleasesSize is the reclaimable size of superseded codex releases.
func staleCodexReleasesSize(home string) int64 {
	var total int64
	for _, p := range staleCodexReleases(home) {
		total += pathSize(p)
	}
	return total
}

// applyStaleCodexReleases deletes every codex release except the newest.
func applyStaleCodexReleases(opts Options) error {
	for _, p := range staleCodexReleases(opts.Home) {
		if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// codexReleaseVersion extracts the semver-ish prefix of a release dir name,
// e.g. "0.163.0-aarch64-apple-darwin" -> "0.163.0".
func codexReleaseVersion(name string) (string, bool) {
	base := name
	if i := strings.IndexByte(base, '-'); i >= 0 {
		base = base[:i]
	}
	parts := strings.Split(base, ".")
	if len(parts) < 2 {
		return "", false
	}
	for _, p := range parts {
		if p == "" {
			return "", false
		}
		for _, r := range p {
			if r < '0' || r > '9' {
				return "", false
			}
		}
	}
	return base, true
}

// compareCodexVersions compares two release dir names by their numeric prefix.
func compareCodexVersions(a, b string) int {
	av, _ := codexReleaseVersion(a)
	bv, _ := codexReleaseVersion(b)
	ap := strings.Split(av, ".")
	bp := strings.Split(bv, ".")
	for i := 0; i < len(ap) && i < len(bp); i++ {
		ai, _ := strconv.Atoi(ap[i])
		bi, _ := strconv.Atoi(bp[i])
		if ai != bi {
			if ai < bi {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(ap) < len(bp):
		return -1
	case len(ap) > len(bp):
		return 1
	}
	return strings.Compare(a, b)
}

// toolMissingError is returned when brew/npm cannot be found.
type toolMissingError struct {
	Tool string
	Err  error
}

func (e *toolMissingError) Error() string {
	return fmt.Sprintf("%s not found: %v", e.Tool, e.Err)
}

// lookTool finds tool on PathEnv (exclusive) when set, else system PATH.
func lookTool(opts Options, tool string) (string, error) {
	if opts.PathEnv != "" {
		// Exclusive: only search PathEnv directories (PATH-style, colon-separated).
		for _, dir := range filepath.SplitList(opts.PathEnv) {
			if dir == "" {
				continue
			}
			candidate := filepath.Join(dir, tool)
			if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
				// Ensure executable bit roughly (unix); still try if present.
				return candidate, nil
			}
		}
		return "", &toolMissingError{Tool: tool, Err: fmt.Errorf("not in PathEnv")}
	}
	path, err := exec.LookPath(tool)
	if err != nil {
		return "", &toolMissingError{Tool: tool, Err: err}
	}
	return path, nil
}

// toolChildEnv builds HOME + PathEnv-first PATH for brew/npm/xcrun children.
func toolChildEnv(opts Options) []string {
	env := []string{
		"HOME=" + opts.Home,
	}
	if opts.PathEnv != "" {
		// PathEnv first (stubs win); keep ambient PATH for coreutils in stubs.
		ambient := os.Getenv("PATH")
		if ambient != "" {
			env = append(env, "PATH="+opts.PathEnv+string(os.PathListSeparator)+ambient)
		} else {
			env = append(env, "PATH="+opts.PathEnv)
		}
	}
	return env
}

// runExternalTool runs brew/npm/xcrun with HOME=opts.Home and PathEnv-first PATH.
// Binary resolution uses lookTool (PathEnv exclusive when set) so isolation
// works; child PATH is PathEnv first + ambient PATH so stub scripts can still
// find helpers like rm/sh.
func runExternalTool(opts Options, tool string, args ...string) error {
	bin, err := lookTool(opts, tool)
	if err != nil {
		return err
	}
	c := exec.Command(bin, args...)
	c.Env = append(os.Environ(), toolChildEnv(opts)...)
	c.Stdout = io.Discard
	c.Stderr = opts.Stderr
	return c.Run()
}

// runExternalToolOutput is like runExternalTool but returns combined child stdout.
func runExternalToolOutput(opts Options, tool string, args ...string) ([]byte, error) {
	bin, err := lookTool(opts, tool)
	if err != nil {
		return nil, err
	}
	env := toolChildEnv(opts)
	var stdout bytes.Buffer
	c := exec.Command(bin, args...)
	c.Env = append(os.Environ(), env...)
	c.Stdout = &stdout
	c.Stderr = opts.Stderr
	if err := c.Run(); err != nil {
		return stdout.Bytes(), err
	}
	return stdout.Bytes(), nil
}

// isToolMissing reports whether err is a missing-tool condition.
func isToolMissing(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*toolMissingError); ok {
		return true
	}
	// Fallback: exec.ErrNotFound wrapped.
	return strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "executable file not found")
}

// warnf prints a warning line on stderr.
func warnf(w io.Writer, format string, args ...interface{}) {
	if w == nil {
		w = os.Stderr
	}
	fmt.Fprintf(w, "warning: "+format+"\n", args...)
}
