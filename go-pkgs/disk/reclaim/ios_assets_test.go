package reclaim

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveIOSSimAssetRoot_default(t *testing.T) {
	t.Parallel()
	root, err := resolveIOSSimAssetRoot(Options{})
	if err != nil {
		t.Fatal(err)
	}
	if root != defaultIOSSimAssetRoot {
		t.Fatalf("got %q", root)
	}
}

func TestResolveIOSSimAssetRoot_customAbsolute(t *testing.T) {
	t.Parallel()
	root, err := resolveIOSSimAssetRoot(Options{IOSSimAssetRoot: "/tmp/fixture-assets"})
	if err != nil {
		t.Fatal(err)
	}
	if root != "/tmp/fixture-assets" {
		t.Fatalf("got %q", root)
	}
}

func TestResolveIOSSimAssetRoot_rejectRelative(t *testing.T) {
	t.Parallel()
	_, err := resolveIOSSimAssetRoot(Options{IOSSimAssetRoot: "relative/path"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRemoveIOSSimAssetsContents(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.dmg"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b"), []byte("xx"), 0o644); err != nil {
		t.Fatal(err)
	}
	if pathSize(dir) == 0 {
		t.Fatal("expected size > 0")
	}
	if err := removeIOSSimAssetsContents(dir); err != nil {
		t.Fatal(err)
	}
	if pathSize(dir) != 0 {
		t.Fatalf("size after wipe = %d", pathSize(dir))
	}
}

func TestApplyIOSSimAssets_writableNoSudo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "blob"), make([]byte, 1000), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	opts := Options{
		IOSSimAssetRoot:     dir,
		IncludeIOSSimAssets: true,
		Geteuid:             func() int { return 0 },
		Stdout:              &stdout,
		Stderr:              &stderr,
	}
	before, after, err := applyIOSSimAssets(opts)
	if err != nil {
		t.Fatal(err)
	}
	if before != 1000 {
		t.Fatalf("before=%d", before)
	}
	if after != 0 {
		t.Fatalf("after=%d", after)
	}
}

func TestApplyIOSSimAssets_needsSudoNonTTY(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "blob"), make([]byte, 500), 0o644); err != nil {
		t.Fatal(err)
	}
	// Make unwritable to this process (best-effort; skip if cannot chmod as owner still writes).
	// Simulate via Geteuid!=0 and canWrite forced false by removing write on dir for others —
	// owner can still write. Use SudoRun not called: inject canWrite failure by using
	// Geteuid!=0 and SudoRun nil and IsTerminal false — but canWrite still true as owner.
	// So inject SudoRun path only when canWrite is false: use a path we can't write.
	// Simpler: Geteuid=1, IsTerminal=false, and override by making remove need sudo via
	// custom: we set SudoRun=nil, Geteuid=1, and use a root-only path that doesn't exist size 0.
	// For non-zero size unwritable: chmod 0555 on dir after creating file.
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	opts := Options{
		IOSSimAssetRoot: dir,
		Geteuid:         func() int { return 501 },
		IsTerminal:      func() bool { return false },
	}
	// Owner may still write on some systems despite 0555 on dir for create-temp.
	// If canWrite still true, skip this assertion.
	if canWriteAssetRoot(dir) {
		t.Skip("owner can still write fixture dir; cannot simulate EPERM")
	}
	_, _, err := applyIOSSimAssets(opts)
	if err == nil {
		t.Fatal("expected privilege error")
	}
	if !strings.Contains(err.Error(), "admin privileges") {
		t.Fatalf("err=%v", err)
	}
}

func TestApplyIOSSimAssets_sudoRunInjected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "blob"), make([]byte, 800), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if canWriteAssetRoot(dir) {
		// Still use SudoRun by pretending not root and forcing sudo path:
		// if can write, apply won't call SudoRun. Force via Geteuid=0 instead for wipe.
		opts := Options{
			IOSSimAssetRoot: dir,
			Geteuid:         func() int { return 0 },
		}
		// chmod back to allow remove as root-inject
		_ = os.Chmod(dir, 0o755)
		before, after, err := applyIOSSimAssets(opts)
		if err != nil {
			t.Fatal(err)
		}
		if before != 800 || after != 0 {
			t.Fatalf("before=%d after=%d", before, after)
		}
		return
	}
	called := false
	opts := Options{
		IOSSimAssetRoot: dir,
		Geteuid:         func() int { return 501 },
		IsTerminal:      func() bool { return true },
		SudoRun: func(o Options) error {
			called = true
			_ = os.Chmod(dir, 0o755)
			return removeIOSSimAssetsContents(dir)
		},
	}
	before, after, err := applyIOSSimAssets(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("SudoRun not called")
	}
	if before != 800 || after != 0 {
		t.Fatalf("before=%d after=%d", before, after)
	}
}

func TestRun_dryRunListsMobileAsset(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	asset := filepath.Join(dir, "assets")
	if err := os.MkdirAll(asset, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(asset, "x.dmg"), make([]byte, 2048), 0o644); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(dir, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	// empty path so xcrun missing
	bin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	res, err := Run(Options{
		Home:            home,
		DryRun:          true,
		PathEnv:         bin,
		IOSSimAssetRoot: asset,
		SkipVolume:      true,
		Stdout:          &stdout,
		Stderr:          &stderr,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, "iOS Simulator MobileAsset") {
		t.Fatalf("missing MobileAsset section:\n%s", out)
	}
	if !strings.Contains(out, "--ios-sim-assets") {
		t.Fatalf("missing assets annotation:\n%s", out)
	}
	if !strings.Contains(out, "not deleted by default") {
		t.Fatalf("missing runtime annotation:\n%s", out)
	}
	_ = res
}

func TestRun_liveIOSSimAssets(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	asset := filepath.Join(dir, "assets")
	if err := os.MkdirAll(asset, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(asset, "x.dmg"), make([]byte, 4096), 0o644); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(dir, "home")
	_ = os.MkdirAll(home, 0o755)
	bin := filepath.Join(dir, "bin")
	_ = os.MkdirAll(bin, 0o755)
	var stdout, stderr bytes.Buffer
	res, err := Run(Options{
		Home:                home,
		IncludeIOSSimAssets: true,
		IOSSimAssetRoot:     asset,
		Geteuid:             func() int { return 0 },
		PathEnv:             bin,
		SkipVolume:          true,
		Stdout:              &stdout,
		Stderr:              &stderr,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalFreed != 4096 {
		t.Fatalf("TotalFreed=%d want 4096\n%s", res.TotalFreed, stdout.String())
	}
	if pathSize(asset) != 0 {
		t.Fatal("asset not wiped")
	}
}

func TestWrapSIPAssetError_includesRecoveryRecipe(t *testing.T) {
	t.Parallel()
	err := wrapSIPAssetError(defaultIOSSimAssetRoot, fmt.Errorf("Operation not permitted"))
	if err == nil {
		t.Fatal("expected error")
	}
	s := err.Error()
	for _, want := range []string{"csrutil disable", "csrutil enable", "com_apple_MobileAsset_iOSSimulatorRuntime", "SIP"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in:\n%s", want, s)
		}
	}
}

func TestPrintIOSSimAssetsSection_includesRecipeWhenSized(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	printIOSSimAssetsSection(&buf, false, defaultIOSSimAssetRoot, 25<<30, false)
	out := buf.String()
	if !strings.Contains(out, "csrutil disable") {
		t.Fatalf("expected Recovery recipe:\n%s", out)
	}
}
