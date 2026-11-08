package reclaim

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFixtureFile creates path with exactly n bytes of 'a'.
func writeFixtureFile(t *testing.T, path string, n int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", path, err)
	}
	if err := os.WriteFile(path, bytes.Repeat([]byte("a"), n), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
}

// writeStubTool installs an executable shell stub named name into dir.
func writeStubTool(t *testing.T, dir, name, script string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
		t.Fatalf("write stub %s: %v", name, err)
	}
}

func selections(opts Options) map[string]bool {
	out := map[string]bool{}
	for _, def := range selectSteps(opts) {
		out[def.ID] = true
	}
	return out
}

func TestSelectSteps_DefaultOnlyDefaultTier(t *testing.T) {
	t.Parallel()
	got := selections(Options{})
	want := []string{
		"sandbox-gocache", "codex-logs", "home-logs", "brew-cleanup",
		"npm-cache", "remote-agent-logs", "podman-applehv-cache", "iterm-bak-apps",
	}
	if len(got) != len(want) {
		t.Fatalf("selected %d steps, want %d: %v", len(got), len(want), got)
	}
	for _, id := range want {
		if !got[id] {
			t.Fatalf("missing default step %s in %v", id, got)
		}
	}
	for _, id := range []string{
		"go-build-cache", "pnpm-store", "npm-npx", "bun-cache",
		"codex-releases", "xcache", stepMSPlaywright, stepOpenCodeSnapshot,
	} {
		if got[id] {
			t.Fatalf("aggressive step %s must not run by default", id)
		}
	}
}

func TestSelectSteps_AggressiveEnablesAll(t *testing.T) {
	t.Parallel()
	got := selections(Options{Aggressive: true})
	for _, id := range []string{
		"go-build-cache", "pnpm-store", "npm-npx", "bun-cache",
		"codex-releases", "xcache", stepMSPlaywright, stepOpenCodeSnapshot,
		"sandbox-gocache", "iterm-bak-apps",
	} {
		if !got[id] {
			t.Fatalf("aggressive run missing step %s in %v", id, got)
		}
	}

	// A single opt-in flag enables only its own step.
	onlyPW := selections(Options{IncludeMSPlaywright: true})
	if !onlyPW[stepMSPlaywright] {
		t.Fatalf("--ms-playwright did not enable its step: %v", onlyPW)
	}
	if onlyPW["xcache"] || onlyPW["go-build-cache"] {
		t.Fatalf("--ms-playwright leaked other aggressive steps: %v", onlyPW)
	}
}

func TestRun_DefaultLeavesAggressiveCachesAlone(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	seed := filepath.Join(home, ".cache", "puppeteer", "chrome")
	writeFixtureFile(t, seed, 1024)
	npx := filepath.Join(home, ".npm", "_npx", "hash", "pkg")
	writeFixtureFile(t, npx, 512)

	var stdout, stderr bytes.Buffer
	res, err := Run(Options{
		Home:       home,
		DryRun:     true,
		Stdout:     &stdout,
		Stderr:     &stderr,
		PathEnv:    t.TempDir(), // no tools available
		SkipVolume: true,
	})
	if err != nil {
		t.Fatalf("run: %v (stderr=%s)", err, stderr.String())
	}
	out := stdout.String()
	for _, id := range []string{"go-build-cache", "pnpm-store", "npm-npx", "bun-cache", "codex-releases", "xcache"} {
		if strings.Contains(out, id) {
			t.Fatalf("default dry-run listed aggressive step %s:\n%s", id, out)
		}
	}
	if res.TotalFreed != 0 {
		t.Fatalf("default dry-run freed = %d, want 0 (aggressive caches not counted)", res.TotalFreed)
	}
	if _, err := os.Stat(seed); err != nil {
		t.Fatalf("default run touched .cache: %v", err)
	}
	if _, err := os.Stat(npx); err != nil {
		t.Fatalf("default run touched .npm/_npx: %v", err)
	}
}

func TestRun_AggressiveCleansCaches(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	stubs := t.TempDir()
	writeStubTool(t, stubs, "go", "#!/bin/sh\nrm -rf \"$HOME/Library/Caches/go-build\"\n")
	writeStubTool(t, stubs, "pnpm", "#!/bin/sh\nrm -rf \"$HOME/Library/pnpm/store\"\n")

	writeFixtureFile(t, filepath.Join(home, "Library", "Caches", "go-build", "aa"), 1000)
	writeFixtureFile(t, filepath.Join(home, "Library", "pnpm", "store", "v3", "x"), 2000)
	writeFixtureFile(t, filepath.Join(home, ".npm", "_npx", "hash", "pkg"), 4000)
	writeFixtureFile(t, filepath.Join(home, ".bun", "install", "cache", "y"), 8000)
	writeFixtureFile(t, filepath.Join(home, ".cache", "puppeteer", "z"), 16000)
	releases := filepath.Join(home, ".codex", "packages", "standalone", "releases")
	oldRelease := filepath.Join(releases, "0.155.0-aarch64-apple-darwin", "bin", "codex")
	newRelease := filepath.Join(releases, "0.163.0-aarch64-apple-darwin", "bin", "codex")
	writeFixtureFile(t, oldRelease, 100)
	writeFixtureFile(t, newRelease, 200)
	// Never touched: the Go module cache is excluded from every tier.
	modCache := filepath.Join(home, "go", "pkg", "mod", "cache", "download", "x.zip")
	writeFixtureFile(t, modCache, 500)
	// A default step so the run also proves default steps still execute.
	writeFixtureFile(t, filepath.Join(home, ".codex", "logs_2.sqlite"), 50)

	var stdout, stderr bytes.Buffer
	res, err := Run(Options{
		Home:       home,
		Aggressive: true,
		Stdout:     &stdout,
		Stderr:     &stderr,
		PathEnv:    stubs,
		SkipVolume: true,
	})
	if err != nil {
		t.Fatalf("run: %v (stderr=%s)", err, stderr.String())
	}

	for _, gone := range []string{
		filepath.Join(home, "Library", "Caches", "go-build"),
		filepath.Join(home, "Library", "pnpm", "store"),
		filepath.Join(home, ".npm", "_npx"),
		filepath.Join(home, ".bun", "install", "cache"),
		filepath.Join(home, ".cache"),
		filepath.Join(home, ".codex", "logs_2.sqlite"),
		filepath.Dir(filepath.Dir(oldRelease)),
	} {
		if _, err := os.Stat(gone); !os.IsNotExist(err) {
			t.Fatalf("%s still exists after aggressive run (err=%v)", gone, err)
		}
	}
	if _, err := os.Stat(newRelease); err != nil {
		t.Fatalf("aggressive run deleted the newest codex release: %v", err)
	}
	if _, err := os.Stat(modCache); err != nil {
		t.Fatalf("aggressive run touched the go module cache: %v", err)
	}

	wantFreed := int64(50 + 1000 + 2000 + 4000 + 8000 + 100 + 16000)
	if res.TotalFreed != wantFreed {
		t.Fatalf("TotalFreed = %d, want %d\nstdout:\n%s", res.TotalFreed, wantFreed, stdout.String())
	}
	byID := map[string]int64{}
	for _, s := range res.Steps {
		byID[s.ID] = s.Freed
	}
	if byID["codex-releases"] != 100 {
		t.Fatalf("codex-releases freed = %d, want 100 (only superseded releases)", byID["codex-releases"])
	}
	// Missing brew/npm only warn.
	if !strings.Contains(stderr.String(), "warning:") {
		t.Fatalf("expected missing-tool warnings for brew/npm, stderr=%q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Total freed:") {
		t.Fatalf("missing live total line:\n%s", stdout.String())
	}
}

func TestRun_MobileAssetFailureWarnsAndContinues(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	assets := filepath.Join(t.TempDir(), "MobileAsset_iOSSimulatorRuntime")
	writeFixtureFile(t, filepath.Join(assets, "signed", "blob"), 500)
	if err := os.Chmod(assets, 0o555); err != nil {
		t.Fatalf("chmod assets: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(assets, 0o755) })

	writeFixtureFile(t, filepath.Join(home, "Library", "Caches", "ms-playwright", "chromium", "x"), 300)
	writeFixtureFile(t, filepath.Join(home, ".local", "share", "opencode", "snapshot", "h", "b"), 400)

	var stdout, stderr bytes.Buffer
	res, err := Run(Options{
		Home:                home,
		Aggressive:          true,
		IncludeIOSSimAssets: true,
		IOSSimAssetRoot:     assets,
		Geteuid:             func() int { return 501 },
		IsTerminal:          func() bool { return false },
		Stdout:              &stdout,
		Stderr:              &stderr,
		PathEnv:             t.TempDir(),
		SkipVolume:          true,
	})
	if err != nil {
		t.Fatalf("MobileAsset failure must not fail the run: %v", err)
	}
	if !strings.Contains(stderr.String(), "warning: ios-sim-mobileasset failed") {
		t.Fatalf("expected warning, stderr=%q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "admin") {
		t.Fatalf("warning should explain privileges, stderr=%q", stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "ms-playwright") || !strings.Contains(out, "opencode-snapshot") {
		t.Fatalf("steps after the failing MobileAsset step did not run:\n%s", out)
	}
	byID := map[string]int64{}
	for _, s := range res.Steps {
		byID[s.ID] = s.Freed
	}
	if byID[stepMSPlaywright] != 300 {
		t.Fatalf("ms-playwright freed = %d, want 300", byID[stepMSPlaywright])
	}
	if byID[stepOpenCodeSnapshot] != 400 {
		t.Fatalf("opencode-snapshot freed = %d, want 400", byID[stepOpenCodeSnapshot])
	}
	if _, err := os.Stat(filepath.Join(assets, "signed", "blob")); err != nil {
		t.Fatalf("blocked MobileAsset step deleted content anyway: %v", err)
	}
	if !strings.Contains(out, "Total freed:") {
		t.Fatalf("report still printed after a blocked step:\n%s", out)
	}
}

func TestStaleCodexReleases_KeepsNewestOnly(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	releases := codexReleasesDir(home)
	for _, name := range []string{
		"0.155.0-aarch64-apple-darwin",
		"0.155.1-aarch64-apple-darwin",
		"0.161.0-aarch64-apple-darwin",
		"0.162.1-aarch64-apple-darwin",
		"0.163.0-aarch64-apple-darwin",
		"not-a-release",
	} {
		if err := os.MkdirAll(filepath.Join(releases, name, "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	stale := staleCodexReleases(home)
	if len(stale) != 4 {
		t.Fatalf("stale = %v, want 4 superseded releases", stale)
	}
	for _, p := range stale {
		if filepath.Base(p) == "0.163.0-aarch64-apple-darwin" {
			t.Fatalf("newest release listed as stale: %v", stale)
		}
		if filepath.Base(p) == "not-a-release" {
			t.Fatalf("non-version dir listed as stale: %v", stale)
		}
	}
	if got := staleCodexReleases(home); got[0] != filepath.Join(releases, "0.155.0-aarch64-apple-darwin") {
		t.Fatalf("stale order = %v, want oldest first", got)
	}
}
