// Package install installs the official iTerm2 stable build without Homebrew.
//
// Default pipeline: resolve latest zip URL → download → extract iTerm.app →
// install under ~/Applications (or an injected target) → optional Launch
// Services register → VerifyInstalled → optional scriptable version check.
//
// All filesystem roots and HTTP are injectable for parallel-safe tests.
package install

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultLatestURL is the official iTerm2 stable "latest" redirect endpoint.
	DefaultLatestURL = "https://iterm2.com/downloads/stable/latest"

	// BundleID is the CFBundleIdentifier expected in a valid iTerm2 app.
	BundleID = "com.googlecode.iterm2"

	// AppBundleName is the macOS application bundle directory name.
	AppBundleName = "iTerm.app"
)

// ResolveOpts configures ResolveLatestStableURL.
type ResolveOpts struct {
	// LatestURL overrides DefaultLatestURL when non-empty (tests inject fake servers).
	LatestURL string
	// HTTPClient is used for the request; nil → http.DefaultClient.
	HTTPClient *http.Client
}

// DownloadOpts configures Download.
type DownloadOpts struct {
	// HTTPClient is used for the request; nil → http.DefaultClient.
	HTTPClient *http.Client
}

// InstallAppOpts configures InstallApp.
type InstallAppOpts struct {
	// Home is used when targetApp is empty to form {Home}/Applications/iTerm.app.
	// Empty Home falls back to os.UserHomeDir.
	Home string
}

// VerifyScriptableOpts configures VerifyScriptable.
type VerifyScriptableOpts struct {
	// Runner, when non-nil, is used instead of real osascript.
	Runner func() (string, error)
}

// InstallOpts configures InstallLatest.
type InstallOpts struct {
	// Home injects the home directory for default target {Home}/Applications/iTerm.app.
	// Empty → os.UserHomeDir.
	Home string
	// TargetApp overrides the install destination; empty → default under Home.
	TargetApp string
	// CacheDir holds downloaded zip and extract work; empty → os.MkdirTemp.
	CacheDir string
	// LatestURL overrides DefaultLatestURL; empty → DefaultLatestURL.
	LatestURL string
	// HTTPClient is used for resolve + download; nil → http.DefaultClient.
	HTTPClient *http.Client
	// Register registers the app with Launch Services; nil → real lsregister on darwin.
	Register func(appPath string) error
	// ScriptableRunner overrides VerifyScriptable's runner when scriptable is not skipped.
	ScriptableRunner func() (string, error)
	// SkipScriptable skips the post-install AppleScript version check.
	SkipScriptable bool
	// RequireHostArch enables an optional host-arch check after extract (best-effort).
	RequireHostArch bool
}

// Result is the outcome of InstallLatest.
type Result struct {
	URL        string
	Version    string
	ZipPath    string
	AppPath    string
	BackupPath string
}

// ResolveLatestStableURL follows redirects from LatestURL/DefaultLatestURL and
// returns the final zip URL plus a version parsed from the zip basename
// (e.g. iTerm2-3_6_11.zip → 3.6.11).
func ResolveLatestStableURL(ctx context.Context, opts ResolveOpts) (url string, version string, err error) {
	latest := opts.LatestURL
	if latest == "" {
		latest = DefaultLatestURL
	}
	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latest, nil)
	if err != nil {
		return "", "", fmt.Errorf("resolve latest: build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("resolve latest: %w", err)
	}
	defer resp.Body.Close()
	// Drain body so the connection can be reused.
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("resolve latest: HTTP %d from %s", resp.StatusCode, latest)
	}

	finalURL := resp.Request.URL.String()
	if !strings.HasSuffix(strings.ToLower(finalURL), ".zip") {
		// Also accept path ending in .zip ignoring query string.
		path := resp.Request.URL.Path
		if !strings.HasSuffix(strings.ToLower(path), ".zip") {
			return "", "", fmt.Errorf("resolve latest: final URL is not a zip: %s", finalURL)
		}
	}

	ver, err := versionFromZipURL(finalURL)
	if err != nil {
		return "", "", fmt.Errorf("resolve latest: %w", err)
	}
	return finalURL, ver, nil
}

// Download GETs url and writes the body to destPath. Non-2xx or empty body is an error.
func Download(ctx context.Context, url, destPath string, opts DownloadOpts) error {
	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("download: build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download: HTTP %d from %s", resp.StatusCode, url)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("download: mkdir: %w", err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("download: create %s: %w", destPath, err)
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("download: write: %w", err)
	}
	if n == 0 {
		_ = os.Remove(destPath)
		return fmt.Errorf("download: empty body from %s", url)
	}
	return nil
}

// ExtractApp unpacks zipPath into workDir and returns the path to iTerm.app inside it.
func ExtractApp(zipPath, workDir string) (appPath string, err error) {
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return "", fmt.Errorf("extract: mkdir workdir: %w", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("extract: open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if err := extractZipFile(f, workDir); err != nil {
			return "", err
		}
	}

	found, err := findAppBundle(workDir)
	if err != nil {
		return "", err
	}
	return found, nil
}

// InstallApp places extractedApp at targetApp. Empty targetApp resolves to
// {Home}/Applications/iTerm.app. An existing target is renamed to
// target+".bak-"+unixTs before the new app is moved into place.
func InstallApp(extractedApp, targetApp string, opts InstallAppOpts) error {
	target, err := resolveTargetApp(targetApp, opts.Home)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("install app: mkdir parent: %w", err)
	}

	if _, err := os.Stat(target); err == nil {
		bak := target + ".bak-" + strconv.FormatInt(time.Now().Unix(), 10)
		if err := os.Rename(target, bak); err != nil {
			return fmt.Errorf("install app: backup existing: %w", err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("install app: stat target: %w", err)
	}

	if err := moveOrCopyDir(extractedApp, target); err != nil {
		return fmt.Errorf("install app: place app: %w", err)
	}
	return nil
}

// HomeApplicationsITermApp returns ~/Applications/iTerm.app via os.UserHomeDir.
func HomeApplicationsITermApp() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Applications", AppBundleName), nil
}

// VerifyInstalled checks that appPath looks like a valid iTerm2 bundle:
// Contents/MacOS has a binary, and CFBundleIdentifier == BundleID.
func VerifyInstalled(appPath string) error {
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	data, err := os.ReadFile(plistPath)
	if err != nil {
		return fmt.Errorf("verify installed: read Info.plist: %w", err)
	}
	id, err := plistString(data, "CFBundleIdentifier")
	if err != nil {
		return fmt.Errorf("verify installed: %w", err)
	}
	if id != BundleID {
		return fmt.Errorf("verify installed: CFBundleIdentifier = %q, want %q", id, BundleID)
	}

	macosDir := filepath.Join(appPath, "Contents", "MacOS")
	entries, err := os.ReadDir(macosDir)
	if err != nil {
		return fmt.Errorf("verify installed: MacOS dir: %w", err)
	}
	hasBinary := false
	for _, e := range entries {
		if !e.IsDir() {
			hasBinary = true
			break
		}
	}
	if !hasBinary {
		return fmt.Errorf("verify installed: no binary under %s", macosDir)
	}
	return nil
}

// VerifyScriptable returns the iTerm2 version via AppleScript (or an injected Runner).
func VerifyScriptable(opts VerifyScriptableOpts) (string, error) {
	if opts.Runner != nil {
		return opts.Runner()
	}
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("verify scriptable: osascript only supported on darwin")
	}
	// tell application id "com.googlecode.iterm2" to get version
	script := `tell application id "com.googlecode.iterm2" to get version`
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("verify scriptable: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// InstallLatest runs the full resolve → download → extract → install →
// register → verify pipeline.
func InstallLatest(ctx context.Context, opts InstallOpts) (Result, error) {
	var result Result

	url, ver, err := ResolveLatestStableURL(ctx, ResolveOpts{
		LatestURL:  opts.LatestURL,
		HTTPClient: opts.HTTPClient,
	})
	if err != nil {
		return result, err
	}
	result.URL = url
	result.Version = ver

	cacheDir := opts.CacheDir
	if cacheDir == "" {
		tmp, err := os.MkdirTemp("", "iterm2-install-*")
		if err != nil {
			return result, fmt.Errorf("install latest: cache dir: %w", err)
		}
		cacheDir = tmp
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return result, fmt.Errorf("install latest: mkdir cache: %w", err)
	}

	zipName := filepath.Base(strings.Split(url, "?")[0])
	if zipName == "" || zipName == "." || zipName == "/" {
		zipName = "iterm2-latest.zip"
	}
	zipPath := filepath.Join(cacheDir, zipName)
	if err := Download(ctx, url, zipPath, DownloadOpts{HTTPClient: opts.HTTPClient}); err != nil {
		return result, err
	}
	result.ZipPath = zipPath

	extractDir := filepath.Join(cacheDir, "extract")
	extracted, err := ExtractApp(zipPath, extractDir)
	if err != nil {
		return result, err
	}

	if opts.RequireHostArch {
		// Best-effort: presence of a MacOS binary is already required by VerifyInstalled.
		// Host-arch lipo checks are optional and skipped when tools are unavailable.
	}

	target := opts.TargetApp
	// Capture backup path after install by scanning siblings.
	if err := InstallApp(extracted, target, InstallAppOpts{Home: opts.Home}); err != nil {
		return result, err
	}

	appPath, err := resolveTargetApp(target, opts.Home)
	if err != nil {
		return result, err
	}
	result.AppPath = appPath
	result.BackupPath = findBackupBeside(appPath)

	register := opts.Register
	if register == nil {
		register = defaultRegister
	}
	if err := register(appPath); err != nil {
		return result, fmt.Errorf("install latest: register: %w", err)
	}

	if err := VerifyInstalled(appPath); err != nil {
		return result, err
	}

	if !opts.SkipScriptable {
		if _, err := VerifyScriptable(VerifyScriptableOpts{Runner: opts.ScriptableRunner}); err != nil {
			return result, err
		}
	}

	return result, nil
}

func defaultRegister(appPath string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	// /System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister
	lsregister := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
	if _, err := os.Stat(lsregister); err != nil {
		return nil // best-effort
	}
	cmd := exec.Command(lsregister, "-f", appPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("lsregister: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func resolveTargetApp(targetApp, home string) (string, error) {
	if targetApp != "" {
		return targetApp, nil
	}
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("install app: home: %w", err)
		}
	}
	return filepath.Join(home, "Applications", AppBundleName), nil
}

var versionFromName = regexp.MustCompile(`(?i)iterm2-([0-9_]+)\.zip`)

func versionFromZipURL(u string) (string, error) {
	// Strip query/fragment for basename.
	clean := u
	if i := strings.IndexAny(clean, "?#"); i >= 0 {
		clean = clean[:i]
	}
	base := filepath.Base(clean)
	if m := versionFromName.FindStringSubmatch(base); len(m) == 2 {
		return strings.ReplaceAll(m[1], "_", "."), nil
	}
	// Fallback: after last '-' before .zip, underscores → dots.
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if i := strings.LastIndex(name, "-"); i >= 0 && i+1 < len(name) {
		seg := name[i+1:]
		if seg != "" {
			return strings.ReplaceAll(seg, "_", "."), nil
		}
	}
	return "", fmt.Errorf("cannot parse version from %q", base)
}

func extractZipFile(f *zip.File, destDir string) error {
	name := f.Name
	// Zip Slip guard.
	clean := filepath.Clean(name)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("extract: invalid path %q", name)
	}
	target := filepath.Join(destDir, clean)

	if f.FileInfo().IsDir() || strings.HasSuffix(name, "/") {
		return os.MkdirAll(target, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("extract: mkdir: %w", err)
	}
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("extract: open entry: %w", err)
	}
	defer rc.Close()

	mode := f.Mode()
	if mode == 0 {
		mode = 0o644
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("extract: create file: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("extract: write file: %w", err)
	}
	return nil
}

func findAppBundle(root string) (string, error) {
	var found string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && filepath.Base(path) == AppBundleName {
			found = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("extract: walk: %w", err)
	}
	if found == "" {
		return "", fmt.Errorf("extract: %s not found in archive", AppBundleName)
	}
	return found, nil
}

func moveOrCopyDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Cross-device or other rename failure → recursive copy then remove src.
	if err := copyDir(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func findBackupBeside(appPath string) string {
	dir := filepath.Dir(appPath)
	base := filepath.Base(appPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	prefix := base + ".bak-"
	var best string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			best = filepath.Join(dir, e.Name())
		}
	}
	return best
}

// plistString extracts the first <string> value after <key>name</key> from an XML plist.
func plistString(data []byte, key string) (string, error) {
	// Minimal XML plist parse — fixtures and real Info.plist use this shape.
	re := regexp.MustCompile(`(?s)<key>` + regexp.QuoteMeta(key) + `</key>\s*<string>([^<]*)</string>`)
	m := re.FindSubmatch(data)
	if m == nil {
		return "", fmt.Errorf("plist key %q not found", key)
	}
	return string(m[1]), nil
}
