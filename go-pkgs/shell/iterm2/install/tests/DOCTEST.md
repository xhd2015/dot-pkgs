# shell/iterm2/install — Official iTerm2 stable install (no Homebrew)

## Version

0.0.2

Classic TDD (greenfield) doctests for package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/install`.

Leaves call the intended public API. The package does **not** exist yet → suite
stays **RED** until the implementer lands source under `shell/iterm2/install/`.

**Default layer: L2** in-process library API with injectable HTTP, Home,
TargetApp, Register, and ScriptableRunner. No real network in the default suite.
Parallel-safe: no `t.Setenv` / `t.Chdir` / process env mutation for isolation.

**Out of scope this cycle:** brew recipes, parent AppleScript open scripts, auto
`brew uninstall`, deleting `/Applications/iTerm.app` when a home install exists,
arch-specific download URLs (stable zip is **universal**).

## DSN (Domain Specific Notion)

### Participants

- **`ResolveLatestStableURL(ctx, opts)`** — follows redirects from
  `opts.LatestURL` when set, else `DefaultLatestURL`
  (`https://iterm2.com/downloads/stable/latest`); returns final zip URL + version
  parsed from filename (`iTerm2-3_6_11.zip` → `3.6.11`). HTTP via injectable
  `opts.HTTPClient` (nil → `http.DefaultClient`).
- **`Download(ctx, url, destPath, opts)`** — GET zip to `destPath` (injectable
  HTTP). Errors on non-2xx or empty body.
- **`ExtractApp(zipPath, workDir)`** — unpacks zip into `workDir`; returns path
  to `iTerm.app` directory inside `workDir`. Errors if bundle missing.
- **`InstallApp(extractedApp, targetApp, opts)`** — places extracted app at
  target. Empty `targetApp` → `{Home}/Applications/iTerm.app` where Home is
  `opts.Home` or `os.UserHomeDir`. Existing target is renamed to
  `target+".bak-"+unixTs` then new app moved/copied into place.
- **`HomeApplicationsITermApp()`** — production helper via `os.UserHomeDir`.
  Tests inject `Home` / `TargetApp` instead of mutating process env.
- **`VerifyInstalled(appPath)`** — checks bundle layout (`Contents/MacOS/…`) and
  `CFBundleIdentifier` = `BundleID` (`com.googlecode.iterm2`).
- **`VerifyScriptable(opts)`** — AppleScript get version via injectable
  `opts.Runner`; nil → real `osascript` on darwin.
- **`InstallLatest(ctx, opts)`** — resolve → download → extract → install →
  optional `Register` (lsregister) → `VerifyInstalled` → optional scriptable.
  Returns `Result{URL, Version, ZipPath, AppPath, BackupPath}`.
  **`InstallOpts.LatestURL`** (empty → `DefaultLatestURL`) is required for L2
  HTTP injection without rewriting the real host.
- **Constants** — `DefaultLatestURL`, `BundleID`, `AppBundleName` (`iTerm.app`).

### Behaviors

**Resolve**

- Redirect chain ends on `*.zip` → return that URL + version from basename
  (`_` → `.` between version digits).
- Final non-zip URL → error.
- HTTP status/transport error → error.
- Stable product URL is universal: resolved URL must **not** contain `arm64` or
  `amd64`.

**Download**

- 2xx with non-empty body → write dest file.
- 404 / empty body → error.

**Extract**

- Zip containing `iTerm.app/Contents/...` → app path under workDir.
- Zip without `iTerm.app` → error.

**InstallApp**

- Fresh target: app at target with verifiable layout.
- Replace: previous tree survives as `*.bak-<unix_ts>`; new app at target.
- Empty target + injected Home → `{Home}/Applications/iTerm.app`.

**VerifyInstalled**

- Mini-bundle with Info.plist BundleID + MacOS binary → nil.
- Missing MacOS binary or wrong BundleID → error.

**VerifyScriptable**

- Injected runner success → version string.
- Injected runner failure → error.

**InstallLatest**

- Fake HTTP serving zip fixture + injected Home + `SkipScriptable` →
  `Result.AppPath` under home Applications; `VerifyInstalled` passes.
- Injected `Register` (no real lsregister). No arch URL branching.

## Decision Tree

```
shell/iterm2/install/tests/
├── DOCTEST.md
├── SETUP.md
├── resolve/                              # ResolveLatestStableURL
│   ├── redirect-to-zip/                  # 302 → iTerm2-3_6_11.zip; version 3.6.11; no arch
│   ├── non-zip-final/                    # final URL not .zip → error
│   └── http-error/                       # HTTP 500 → error
├── download/                             # Download
│   ├── happy-path/                       # writes zip bytes to DestPath
│   ├── http-404/                         # 404 → error
│   └── empty-body/                       # 200 empty body → error
├── extract/                              # ExtractApp
│   ├── valid-app-bundle/                 # zip with iTerm.app → app path
│   └── missing-iterm-app/                # zip without iTerm.app → error
├── install-app/                          # InstallApp
│   ├── fresh-install/                    # new target under case dir
│   ├── replace-with-backup/              # existing → .bak-<ts>; new app in place
│   └── default-home-applications/        # empty target + Home → ~/Applications/iTerm.app
├── verify-installed/                     # VerifyInstalled
│   ├── good-mini-bundle/                 # Info.plist + MacOS binary → ok
│   ├── missing-macos-binary/             # no Contents/MacOS binary → error
│   └── wrong-bundle-id/                  # CFBundleIdentifier ≠ BundleID → error
├── verify-scriptable/                    # VerifyScriptable (injected runner)
│   ├── runner-success/                   # runner returns version
│   └── runner-fail/                      # runner error → error
└── install-latest/                       # InstallLatest orchestration
    └── pipeline-fake-http/               # fake HTTP zip + Home; SkipScriptable
```

### Parameter significance (high → low)

1. **Operation** — resolve | download | extract | install-app | verify | install-latest
2. **Outcome class** — happy path vs error / replace vs fresh
3. **Injection surface** — HTTP fixture, Home/Target, runner, Register
4. **Fixture shape** — zip contents, Info.plist BundleID, MacOS binary presence

## Test Index

| # | Leaf | API | Description | Classic |
|---|------|-----|-------------|---------|
| 1 | `resolve/redirect-to-zip` | ResolveLatestStableURL | Redirect to `iTerm2-3_6_11.zip` → URL + `3.6.11`; no arm64/amd64 | RED |
| 2 | `resolve/non-zip-final` | ResolveLatestStableURL | Final non-zip URL → error | RED |
| 3 | `resolve/http-error` | ResolveLatestStableURL | HTTP 500 → error | RED |
| 4 | `download/happy-path` | Download | Writes fixture zip to DestPath | RED |
| 5 | `download/http-404` | Download | 404 → error | RED |
| 6 | `download/empty-body` | Download | Empty body → error | RED |
| 7 | `extract/valid-app-bundle` | ExtractApp | Zip with mini iTerm.app → app path | RED |
| 8 | `extract/missing-iterm-app` | ExtractApp | Zip without iTerm.app → error | RED |
| 9 | `install-app/fresh-install` | InstallApp | Fresh install to injected target | RED |
| 10 | `install-app/replace-with-backup` | InstallApp | Existing target backed up as `.bak-<ts>` | RED |
| 11 | `install-app/default-home-applications` | InstallApp | Empty target → `{Home}/Applications/iTerm.app` | RED |
| 12 | `verify-installed/good-mini-bundle` | VerifyInstalled | Good mini-bundle → nil | RED |
| 13 | `verify-installed/missing-macos-binary` | VerifyInstalled | Missing MacOS binary → error | RED |
| 14 | `verify-installed/wrong-bundle-id` | VerifyInstalled | Wrong CFBundleIdentifier → error | RED |
| 15 | `verify-scriptable/runner-success` | VerifyScriptable | Injected runner returns version | RED |
| 16 | `verify-scriptable/runner-fail` | VerifyScriptable | Injected runner fails → error | RED |
| 17 | `install-latest/pipeline-fake-http` | InstallLatest | Full pipeline with fake HTTP zip under Home Applications | RED |

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/iterm2/install/tests
doctest test ./shell/iterm2/install/tests
doctest test -v ./shell/iterm2/install/tests/resolve/redirect-to-zip
doctest test -v ./shell/iterm2/install/tests/install-latest/pipeline-fake-http
```

Classic TDD: expect **RED** (compile failure until package exists, then assert
failures against incomplete implementations) until implementer lands the API.

```go
import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/install"
)

// Request is filled root→leaf. Operation selects which public API Run calls.
type Request struct {
	Operation string // resolve | download | extract | install-app | verify-installed | verify-scriptable | install-latest

	// WorkDir is an isolated temp root (root Setup). Home defaults to WorkDir/home.
	WorkDir string
	Home    string

	// HTTP fixture mode for resolve/download/install-latest.
	// Values: redirect-zip | non-zip | http-error | download-ok | download-404 | download-empty | install-latest
	HTTPMode string

	// Optional override for resolve/install-latest LatestURL (absolute).
	// Empty → fake server "/stable/latest".
	LatestURL string

	// Download
	DownloadURL string
	DestPath    string

	// Extract
	ZipPath        string
	ExtractWorkDir string

	// InstallApp
	ExtractedApp     string
	TargetApp        string
	UseDefaultTarget bool // force targetApp "" so product defaults to Home/Applications

	// VerifyInstalled
	AppPath string

	// VerifyScriptable
	ScriptableVersion string
	ScriptableFail    bool

	// InstallLatest
	SkipScriptable  bool
	RequireHostArch bool
	CacheDir        string

	// Fixture flags
	SeedExistingTarget bool
	ExistingMarker     string // content written into existing target MacOS binary
	BundleIDOverride   string // empty → install.BundleID
	OmitMacOSBinary    bool
	ZipWithoutApp      bool
	FinalZipName       string // default iTerm2-3_6_11.zip
	RegisterShouldFail bool
}

// Response observes API outputs. Filesystem assertions also use WorkDir/Home.
type Response struct {
	URL            string
	Version        string
	AppPath        string
	ZipPath        string
	BackupPath     string
	RegisteredPath string
	RegisterCalls  int
	DestSize       int64
	ResolvedURL    string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d

	if req.WorkDir == "" {
		t.Fatal("WorkDir not set by Setup")
	}
	if req.Home == "" {
		req.Home = filepath.Join(req.WorkDir, "home")
	}
	if err := os.MkdirAll(req.Home, 0o755); err != nil {
		return nil, err
	}

	ctx := context.Background()
	resp := &Response{}

	switch req.Operation {
	case "resolve":
		srv, client := startHTTPFixture(t, req)
		defer srv.Close()
		latest := req.LatestURL
		if latest == "" {
			latest = srv.URL + "/stable/latest"
		}
		url, ver, err := install.ResolveLatestStableURL(ctx, install.ResolveOpts{
			LatestURL:  latest,
			HTTPClient: client,
		})
		resp.URL = url
		resp.ResolvedURL = url
		resp.Version = ver
		return resp, err

	case "download":
		srv, client := startHTTPFixture(t, req)
		defer srv.Close()
		url := req.DownloadURL
		if url == "" {
			url = srv.URL + "/file.zip"
		}
		dest := req.DestPath
		if dest == "" {
			dest = filepath.Join(req.WorkDir, "out", "download.zip")
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return nil, err
		}
		err := install.Download(ctx, url, dest, install.DownloadOpts{HTTPClient: client})
		resp.ZipPath = dest
		if st, stErr := os.Stat(dest); stErr == nil {
			resp.DestSize = st.Size()
		}
		return resp, err

	case "extract":
		zipPath := req.ZipPath
		if zipPath == "" {
			zipPath = writeZipFixture(t, req)
		}
		work := req.ExtractWorkDir
		if work == "" {
			work = filepath.Join(req.WorkDir, "extract")
		}
		if err := os.MkdirAll(work, 0o755); err != nil {
			return nil, err
		}
		appPath, err := install.ExtractApp(zipPath, work)
		resp.AppPath = appPath
		resp.ZipPath = zipPath
		return resp, err

	case "install-app":
		extracted := req.ExtractedApp
		if extracted == "" {
			extracted = writeMiniApp(t, filepath.Join(req.WorkDir, "src", install.AppBundleName), miniAppOpts{
				BundleID: effectiveBundleID(req),
			})
		}
		target := req.TargetApp
		if req.UseDefaultTarget {
			target = ""
		} else if target == "" {
			target = filepath.Join(req.WorkDir, "dest", install.AppBundleName)
		}
		if req.SeedExistingTarget {
			seedPath := target
			if seedPath == "" {
				seedPath = filepath.Join(req.Home, "Applications", install.AppBundleName)
			}
			_ = writeMiniApp(t, seedPath, miniAppOpts{
				BundleID:     effectiveBundleID(req),
				MacOSContent: req.ExistingMarker,
			})
		}
		err := install.InstallApp(extracted, target, install.InstallAppOpts{Home: req.Home})
		if target == "" {
			resp.AppPath = filepath.Join(req.Home, "Applications", install.AppBundleName)
		} else {
			resp.AppPath = target
		}
		resp.BackupPath = findBackupBeside(resp.AppPath)
		return resp, err

	case "verify-installed":
		app := req.AppPath
		if app == "" {
			app = writeMiniApp(t, filepath.Join(req.WorkDir, install.AppBundleName), miniAppOpts{
				BundleID:        effectiveBundleID(req),
				OmitMacOSBinary: req.OmitMacOSBinary,
			})
		}
		resp.AppPath = app
		return resp, install.VerifyInstalled(app)

	case "verify-scriptable":
		ver, err := install.VerifyScriptable(install.VerifyScriptableOpts{
			Runner: func() (string, error) {
				if req.ScriptableFail {
					return "", fmt.Errorf("injected scriptable failure")
				}
				v := req.ScriptableVersion
				if v == "" {
					v = "3.6.11"
				}
				return v, nil
			},
		})
		resp.Version = ver
		return resp, err

	case "install-latest":
		srv, client := startHTTPFixture(t, req)
		defer srv.Close()
		cache := req.CacheDir
		if cache == "" {
			cache = filepath.Join(req.WorkDir, "cache")
		}
		latest := req.LatestURL
		if latest == "" {
			latest = srv.URL + "/stable/latest"
		}
		var registered string
		var regCalls int
		result, err := install.InstallLatest(ctx, install.InstallOpts{
			Home:            req.Home,
			TargetApp:       req.TargetApp,
			CacheDir:        cache,
			LatestURL:       latest,
			HTTPClient:      client,
			SkipScriptable:  true, // default suite never hits real osascript
			RequireHostArch: req.RequireHostArch,
			Register: func(appPath string) error {
				regCalls++
				registered = appPath
				if req.RegisterShouldFail {
					return fmt.Errorf("injected register failure")
				}
				return nil
			},
		})
		resp.RegisteredPath = registered
		resp.RegisterCalls = regCalls
		if err == nil {
			resp.URL = result.URL
			resp.Version = result.Version
			resp.ZipPath = result.ZipPath
			resp.AppPath = result.AppPath
			resp.BackupPath = result.BackupPath
		}
		return resp, err

	default:
		return nil, fmt.Errorf("unknown Operation %q", req.Operation)
	}
}

func effectiveBundleID(req *Request) string {
	if req.BundleIDOverride != "" {
		return req.BundleIDOverride
	}
	return install.BundleID
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
			// pick any; usually one
			best = filepath.Join(dir, e.Name())
		}
	}
	return best
}

type miniAppOpts struct {
	BundleID        string
	OmitMacOSBinary bool
	MacOSContent    string
}

func writeMiniApp(t *testing.T, appPath string, opts miniAppOpts) string {
	t.Helper()
	if opts.BundleID == "" {
		opts.BundleID = install.BundleID
	}
	contents := filepath.Join(appPath, "Contents")
	macos := filepath.Join(contents, "MacOS")
	if err := os.MkdirAll(macos, 0o755); err != nil {
		t.Fatal(err)
	}
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>%s</string>
	<key>CFBundleName</key>
	<string>iTerm2</string>
	<key>CFBundleExecutable</key>
	<string>iTerm2</string>
</dict>
</plist>
`, opts.BundleID)
	if err := os.WriteFile(filepath.Join(contents, "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}
	if !opts.OmitMacOSBinary {
		content := opts.MacOSContent
		if content == "" {
			content = "#!/bin/sh\n# mini iTerm2 placeholder\n"
		}
		if err := os.WriteFile(filepath.Join(macos, "iTerm2"), []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return appPath
}

func writeZipFixture(t *testing.T, req *Request) string {
	t.Helper()
	zipPath := filepath.Join(req.WorkDir, "fixtures", "iterm.zip")
	if err := os.MkdirAll(filepath.Dir(zipPath), 0o755); err != nil {
		t.Fatal(err)
	}
	staging := filepath.Join(req.WorkDir, "fixtures", "staging")
	_ = os.RemoveAll(staging)
	if req.ZipWithoutApp {
		if err := os.MkdirAll(filepath.Join(staging, "not-iterm"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(staging, "not-iterm", "readme.txt"), []byte("no app"), 0o644); err != nil {
			t.Fatal(err)
		}
	} else {
		writeMiniApp(t, filepath.Join(staging, install.AppBundleName), miniAppOpts{
			BundleID:        effectiveBundleID(req),
			OmitMacOSBinary: req.OmitMacOSBinary,
		})
	}
	if err := zipDir(staging, zipPath); err != nil {
		t.Fatal(err)
	}
	return zipPath
}

func zipDir(srcDir, zipPath string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		name := filepath.ToSlash(rel)
		if info.IsDir() {
			_, err := zw.Create(name + "/")
			return err
		}
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	})
}

func startHTTPFixture(t *testing.T, req *Request) (*httptest.Server, *http.Client) {
	t.Helper()
	zipName := req.FinalZipName
	if zipName == "" {
		zipName = "iTerm2-3_6_11.zip"
	}

	var zipBytes []byte
	needZip := req.HTTPMode == "redirect-zip" ||
		req.HTTPMode == "download-ok" ||
		req.HTTPMode == "install-latest" ||
		req.HTTPMode == ""
	if needZip {
		zipPath := writeZipFixture(t, req)
		b, err := os.ReadFile(zipPath)
		if err != nil {
			t.Fatal(err)
		}
		zipBytes = b
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stable/latest", func(w http.ResponseWriter, r *http.Request) {
		switch req.HTTPMode {
		case "http-error":
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		case "non-zip":
			http.Redirect(w, r, "/downloads/not-a-zip.html", http.StatusFound)
			return
		default:
			http.Redirect(w, r, "/downloads/"+zipName, http.StatusFound)
		}
	})
	mux.HandleFunc("/downloads/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/downloads/")
		if strings.HasSuffix(name, ".html") || req.HTTPMode == "non-zip" {
			w.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(w, "<html>not a zip</html>")
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename="+name)
		_, _ = w.Write(zipBytes)
	})
	mux.HandleFunc("/file.zip", func(w http.ResponseWriter, r *http.Request) {
		switch req.HTTPMode {
		case "download-404":
			http.NotFound(w, r)
			return
		case "download-empty":
			w.WriteHeader(http.StatusOK)
			return
		default:
			if len(zipBytes) == 0 {
				zipPath := writeZipFixture(t, req)
				b, err := os.ReadFile(zipPath)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				zipBytes = b
			}
			w.Header().Set("Content-Type", "application/zip")
			_, _ = w.Write(zipBytes)
		}
	})

	srv := httptest.NewServer(mux)
	client := srv.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	}
	return srv, client
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %#v, want %#v", field, got, want)
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q", got, want)
	}
}

func assertNoArchInURL(t *testing.T, url string) {
	t.Helper()
	lower := strings.ToLower(url)
	if strings.Contains(lower, "arm64") || strings.Contains(lower, "amd64") {
		t.Fatalf("resolved URL must be universal (no arch): %q", url)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected dir %s: %v", path, err)
	}
	if !st.IsDir() {
		t.Fatalf("expected dir %s, got non-dir", path)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path %s: %v", path, err)
	}
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	return string(b)
}
```
