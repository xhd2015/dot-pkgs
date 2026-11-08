package reclaim

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/disk/elevate"
)

// defaultIOSSimAssetRoot is the only path wiped by --ios-sim-assets (when not overridden for tests).
const defaultIOSSimAssetRoot = "/System/Library/AssetsV2/com_apple_MobileAsset_iOSSimulatorRuntime"

// resolveIOSSimAssetRoot returns the absolute asset root to measure/delete.
// Empty Options.IOSSimAssetRoot → default system path (must match allowlist exactly).
// Non-empty → injectable test root (must be absolute); not required to equal the default.
func resolveIOSSimAssetRoot(opts Options) (string, error) {
	root := strings.TrimSpace(opts.IOSSimAssetRoot)
	if root == "" {
		return defaultIOSSimAssetRoot, nil
	}
	root = filepath.Clean(root)
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("ios sim asset root must be absolute: %q", root)
	}
	// Refuse empty after clean (e.g. ".") when custom.
	if root == "." || root == string(filepath.Separator) {
		return "", fmt.Errorf("refusing unsafe ios sim asset root: %q", root)
	}
	return root, nil
}

// measureIOSSimAssets returns total size of the MobileAsset runtime cache tree.
func measureIOSSimAssets(opts Options) (root string, size int64, err error) {
	root, err = resolveIOSSimAssetRoot(opts)
	if err != nil {
		return "", 0, err
	}
	return root, pathSize(root), nil
}

func geteuid(opts Options) int {
	if opts.Geteuid != nil {
		return opts.Geteuid()
	}
	return os.Geteuid()
}

func isInteractiveTTY(opts Options) bool {
	if opts.IsTerminal != nil {
		return opts.IsTerminal()
	}
	return elevate.IsTerminal()
}

// canWriteAssetRoot reports whether the current process may delete under root.
func canWriteAssetRoot(root string) bool { return elevate.CanWriteDir(root) }

func removeIOSSimAssetsContents(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		p := filepath.Join(root, e.Name())
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	return nil
}

// applyIOSSimAssets deletes the MobileAsset cache when permitted, elevating via sudo if needed.
// Returns before/after sizes. Fatal privilege errors are returned to the caller (exit 1).
func applyIOSSimAssets(opts Options) (before, after int64, err error) {
	root, err := resolveIOSSimAssetRoot(opts)
	if err != nil {
		return 0, 0, err
	}
	before = pathSize(root)
	if before == 0 {
		// Nothing to free (missing or empty).
		return 0, 0, nil
	}

	// Already privileged or writable (test fixtures / rare non-SIP paths).
	if geteuid(opts) == 0 || canWriteAssetRoot(root) {
		if err := removeIOSSimAssetsContents(root); err != nil {
			return before, pathSize(root), wrapSIPAssetError(root, err)
		}
		after = pathSize(root)
		if after > 0 && after == before {
			// Root but SIP blocked silent partial failure.
			return before, after, wrapSIPAssetError(root, fmt.Errorf("files still present after rm (SIP restricted)"))
		}
		return before, after, nil
	}

	// Need elevation.
	if opts.SudoRun != nil {
		if err := opts.SudoRun(opts); err != nil {
			return before, pathSize(root), err
		}
		return before, pathSize(root), nil
	}

	if !isInteractiveTTY(opts) {
		return before, before, fmt.Errorf("--ios-sim-assets requires admin privileges; re-run under sudo or in a terminal")
	}

	if err := sudoRemoveIOSSimAssets(opts, root); err != nil {
		return before, pathSize(root), err
	}
	return before, pathSize(root), nil
}

// sudoRemoveIOSSimAssets runs interactive sudo to clear restricted flags then rm.
// On modern macOS this often still fails under SIP for AssetsV2 — see wrapSIPAssetError.
func sudoRemoveIOSSimAssets(opts Options, root string) error {
	// Double-check allowlist for default path; custom roots (tests) never reach here if SudoRun set.
	if opts.IOSSimAssetRoot == "" && root != defaultIOSSimAssetRoot {
		return fmt.Errorf("refusing sudo delete of unexpected path %q", root)
	}
	// chflags may clear UF_RESTRICTED when SIP allows; then rm. Both run as root via sudo.
	// Single shell keeps one password prompt.
	script := fmt.Sprintf(
		`set -e
path=%q
/usr/bin/chflags -R nouchg,noschg,norestricted "$path" 2>/dev/null || true
/bin/rm -rf -- "$path"
`, root)
	cmd := exec.Command("sudo", "-p", "Password for clear-tmp --ios-sim-assets: ", "--", "/bin/bash", "-c", script)
	// Attach controlling TTY for password prompt when possible.
	if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		cmd.Stdin = tty
		defer tty.Close()
	} else {
		cmd.Stdin = os.Stdin
	}
	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	warnf(opts.Stderr, "ios-sim-mobileasset requires admin privileges; prompting for sudo password")
	warnf(opts.Stderr, "note: macOS SIP may still block delete under /System even as root")
	if err := cmd.Run(); err != nil {
		return wrapSIPAssetError(root, err)
	}
	// Confirm gone; SIP can make rm "succeed" partially while files remain.
	if pathSize(root) > 0 {
		return wrapSIPAssetError(root, fmt.Errorf("files still present after sudo rm (SIP restricted)"))
	}
	return nil
}

// sipIOSSimAssetRecipe is the only reliable way to remove orphaned MobileAsset
// iOSSimulatorRuntime packs when SIP marks them restricted (sudo rm fails).
const sipIOSSimAssetRecipe = `macOS SIP blocks deleting this folder from a normal session (even as root).
sudo password OK does not help: files are "restricted" under /System/.../AssetsV2.

Try first (may not list orphans after simctl delete):
  • Xcode → Settings → Platforms — delete if anything still listed
  • System Settings → General → Storage — Developer / Xcode if offered

Only reliable full delete (security tradeoff — re-enable SIP immediately after):

  1. Shut down Mac. Boot into Recovery
     (Apple silicon: hold power button until startup options)
  2. Utilities → Terminal, then:
       csrutil disable
  3. Reboot into normal macOS
  4. In Terminal (ONLY this directory — do not wipe all of AssetsV2):
       sudo rm -rf /System/Library/AssetsV2/com_apple_MobileAsset_iOSSimulatorRuntime
  5. Boot Recovery again:
       csrutil enable
  6. Reboot. Confirm:
       du -sh /System/Library/AssetsV2/com_apple_MobileAsset_iOSSimulatorRuntime
       df -h /System/Volumes/Data

There is no supported public CLI that purges these orphans while SIP stays on.
Only mobileassetd is sandboxed to unlink under AssetsV2; userland cannot.`

// wrapSIPAssetError explains Operation not permitted under SIP for MobileAsset paths.
func wrapSIPAssetError(root string, err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if !strings.Contains(msg, "Operation not permitted") &&
		!strings.Contains(msg, "not permitted") &&
		!strings.Contains(msg, "SIP") &&
		!strings.Contains(msg, "still present") {
		return fmt.Errorf("sudo remove ios-sim-mobileasset: %w", err)
	}
	return fmt.Errorf("cannot remove iOS Simulator MobileAsset cache at %s\n\n%s\n\nUnderlying error: %v",
		root, sipIOSSimAssetRecipe, err)
}

// printIOSSimAssetsSection prints MobileAsset inventory for dry-run / reporting.
func printIOSSimAssetsSection(w io.Writer, use bool, root string, size int64, included bool) {
	fmt.Fprintln(w, gray(use, "iOS Simulator MobileAsset:"))
	fmt.Fprintf(w, "  path: %s\n", root)
	fmt.Fprintf(w, "  size: %s (%d B)\n", humanSize(size), size)
	if included {
		fmt.Fprintln(w, "  [would try remove via --ios-sim-assets; SIP often blocks even with sudo]")
	} else {
		fmt.Fprintln(w, "  note: not removed unless --ios-sim-assets; SIP often blocks even with sudo")
	}
	// When the system path has real bulk, always show how to finish cleanup.
	if size > 0 && (root == defaultIOSSimAssetRoot || strings.Contains(root, "com_apple_MobileAsset_iOSSimulatorRuntime")) {
		fmt.Fprintln(w, gray(use, "  if --ios-sim-assets fails with Operation not permitted:"))
		for _, line := range strings.Split(sipIOSSimAssetRecipe, "\n") {
			if line == "" {
				fmt.Fprintln(w)
				continue
			}
			fmt.Fprintln(w, gray(use, "    "+line))
		}
	}
}
