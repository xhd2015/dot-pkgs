package iterm2

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEscapePathForAppleScript(t *testing.T) {
	got := EscapePathForAppleScript(`/tmp/"proj"`)
	want := `/tmp/\"proj\"`
	if got != want {
		t.Fatalf("EscapePathForAppleScript() = %q, want %q", got, want)
	}
}

func TestTellApplicationHeaderQuotedPath(t *testing.T) {
	// Empty → bare name fallback (loads iTerm dictionary by app name).
	if got := TellApplicationHeader(""); got != `tell application "iTerm2"` {
		t.Fatalf("empty path header = %q, want bare iTerm2", got)
	}

	// Non-empty → string-literal path (not POSIX file expression).
	// Expression form fails to compile iTerm terms at osascript parse time.
	path := `/Users/me/Applications/iTerm.app`
	got := TellApplicationHeader(path)
	want := `tell application "/Users/me/Applications/iTerm.app"`
	if got != want {
		t.Fatalf("path header = %q, want %q", got, want)
	}
	if strings.Contains(got, "POSIX file") {
		t.Fatalf("path header must not use POSIX file expression: %q", got)
	}

	// Escapes quotes inside path.
	quoted := TellApplicationHeader(`/tmp/"App".app`)
	if quoted != `tell application "/tmp/\"App\".app"` {
		t.Fatalf("escaped path header = %q", quoted)
	}
}

func TestBuildScriptCreatesWindowAndCDs(t *testing.T) {
	script := BuildScript("/tmp/proj")
	if !strings.Contains(script, "create window with default profile") {
		t.Fatalf("script must create a window: %q", script)
	}
	if !strings.Contains(script, "create tab with default profile") {
		t.Fatal("script must support tab reuse branch")
	}
	if !strings.Contains(script, `set targetDir to "/tmp/proj"`) {
		t.Fatalf("script missing targetDir: %q", script)
	}
	if !strings.Contains(script, `write text ("cd " & quoted form of targetDir)`) {
		t.Fatalf("script must cd via quoted form of targetDir: %q", script)
	}
	if strings.Contains(script, "exec $SHELL") {
		t.Fatal("script must not replace the login shell with a one-shot command")
	}
}

func TestBuildScriptHandlesSpacesAndSingleQuotes(t *testing.T) {
	script := BuildScript("/tmp/my proj/O'Brien")
	if !strings.Contains(script, `set targetDir to "/tmp/my proj/O'Brien"`) {
		t.Fatalf("script missing spaced path: %q", script)
	}
}

func forceDarwin(t *testing.T) {
	t.Helper()
	SetGOOSForTest("darwin")
	t.Cleanup(func() { SetGOOSForTest("") })
}

func TestOpenConfigRejectsNonDirectory(t *testing.T) {
	forceDarwin(t)
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := OpenConfig(file, &Config{
		Installed: func() bool { return true },
		Osascript: func(string) error { return nil },
	}); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("OpenConfig() error = %v, want not a directory", err)
	}
}

func TestOpenConfigUsesInjectedOsascript(t *testing.T) {
	forceDarwin(t)
	dir := t.TempDir()
	var gotScript string
	err := OpenConfig(dir, &Config{
		Installed: func() bool { return true },
		Osascript: func(script string) error {
			gotScript = script
			return nil
		},
	})
	if err != nil {
		t.Fatalf("OpenConfig() error = %v", err)
	}
	if gotScript == "" {
		t.Fatal("expected osascript to be called")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotScript, abs) {
		t.Fatalf("script missing dir %q: %q", abs, gotScript)
	}
}

func TestOpenConfigNotInstalled(t *testing.T) {
	forceDarwin(t)
	err := OpenConfig(t.TempDir(), &Config{
		Installed: func() bool { return false },
	})
	if !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("OpenConfig() error = %v, want ErrNotInstalled", err)
	}
}

func TestResolveAppPathAmongPrefersHomeThenSystem(t *testing.T) {
	root := t.TempDir()
	homeApp := filepath.Join(root, "home", "Applications", "iTerm.app")
	sysApp := filepath.Join(root, "system", "Applications", "iTerm.app")

	// neither
	if got := resolveAppPathAmong([]string{homeApp, sysApp}); got != "" {
		t.Fatalf("resolveAppPathAmong(none) = %q, want empty", got)
	}

	// only system
	if err := os.MkdirAll(sysApp, 0755); err != nil {
		t.Fatal(err)
	}
	if got := resolveAppPathAmong([]string{homeApp, sysApp}); got != sysApp {
		t.Fatalf("resolveAppPathAmong(system only) = %q, want %q", got, sysApp)
	}

	// both: home wins
	if err := os.MkdirAll(homeApp, 0755); err != nil {
		t.Fatal(err)
	}
	if got := resolveAppPathAmong([]string{homeApp, sysApp}); got != homeApp {
		t.Fatalf("resolveAppPathAmong(both) = %q, want home %q", got, homeApp)
	}

	// only home
	if err := os.RemoveAll(sysApp); err != nil {
		t.Fatal(err)
	}
	if got := resolveAppPathAmong([]string{homeApp, sysApp}); got != homeApp {
		t.Fatalf("resolveAppPathAmong(home only) = %q, want %q", got, homeApp)
	}

	// file is not a bundle directory
	filePath := filepath.Join(root, "not-a-bundle")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := resolveAppPathAmong([]string{filePath}); got != "" {
		t.Fatalf("resolveAppPathAmong(file) = %q, want empty", got)
	}
}

func TestAppCandidatesOrder(t *testing.T) {
	cands := appCandidates()
	if len(cands) < 1 {
		t.Fatal("appCandidates() empty")
	}
	// last is always system AppPath
	if cands[len(cands)-1] != AppPath {
		t.Fatalf("last candidate = %q, want %q", cands[len(cands)-1], AppPath)
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		if len(cands) != 1 {
			t.Fatalf("without home, want only AppPath, got %v", cands)
		}
		return
	}
	wantHome := filepath.Join(home, "Applications", "iTerm.app")
	if len(cands) != 2 || cands[0] != wantHome || cands[1] != AppPath {
		t.Fatalf("appCandidates() = %v, want [%q, %q]", cands, wantHome, AppPath)
	}
}

func TestIsInstalledMatchesResolveAppPath(t *testing.T) {
	if IsInstalled() != (ResolveAppPath() != "") {
		t.Fatalf("IsInstalled()=%v but ResolveAppPath()=%q", IsInstalled(), ResolveAppPath())
	}
}

func TestOpenUnsupportedPlatform(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("darwin always supports iterm2 package entry checks")
	}
	err := Open("/tmp")
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("Open() error = %v, want ErrUnsupportedPlatform", err)
	}
}

func TestAppBundleDir(t *testing.T) {
	cases := []struct {
		name string
		exe  string
		want string
	}{
		{
			name: "standard",
			exe:  "/Applications/iTerm.app/Contents/MacOS/iTerm2",
			want: "/Applications/iTerm.app",
		},
		{
			name: "bak variant",
			exe:  "/Users/x/Applications/iTerm.app.bak-123/Contents/MacOS/iTerm2",
			want: "/Users/x/Applications/iTerm.app.bak-123",
		},
		{
			name: "no app ancestor",
			exe:  "/usr/bin/bash",
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := appBundleDir(tc.exe)
			if got != tc.want {
				t.Fatalf("appBundleDir(%q) = %q, want %q", tc.exe, got, tc.want)
			}
		})
	}
}

func TestRunningITermApps_DiskPathsOnly(t *testing.T) {
	// No running processes; both disk paths exist
	tmp := t.TempDir()
	homeApp := filepath.Join(tmp, "Applications", "iTerm.app")
	mkAppDir(t, homeApp)
	sysApp := filepath.Join(tmp, "System", "iTerm.app")
	mkAppDir(t, sysApp)

	got := RunningITermAppsWithOpts(RunningITermAppOpts{
		Getenv:    func(string) string { return "" },
		Home:      func() string { return tmp },
		IsApp:      func(p string) bool {
			return p == homeApp || p == sysApp
		},
		Pids:      func() (string, error) { return "", nil },
		SystemApp: sysApp,
	})

	if len(got) != 2 {
		t.Fatalf("got %v, want 2 paths", got)
	}
	// Home first, then system
	if got[0] != homeApp {
		t.Fatalf("first = %q, want home %q", got[0], homeApp)
	}
	if got[1] != sysApp {
		t.Fatalf("second = %q, want system %q", got[1], sysApp)
	}
}

func TestRunningITermApps_BakVariantRunning(t *testing.T) {
	tmp := t.TempDir()
	homeApp := filepath.Join(tmp, "Applications", "iTerm.app")
	mkAppDir(t, homeApp)

	bakExe := filepath.Join(tmp, "Applications", "iTerm.app.bak-999", "Contents", "MacOS", "iTerm2")
	psOutput := bakExe + "\n"

	got := RunningITermAppsWithOpts(RunningITermAppOpts{
		Getenv: func(string) string { return "" },
		Home:   func() string { return tmp },
		IsApp: func(p string) bool {
			return p == homeApp // only home exists on disk
		},
		Pids: func() (string, error) { return psOutput, nil },
	})

	// Should contain both the home disk path and the bak running path
	found := map[string]bool{}
	for _, p := range got {
		found[p] = true
	}
	if !found[homeApp] {
		t.Fatalf("missing home app %q in %v", homeApp, got)
	}
	bakApp := filepath.Join(tmp, "Applications", "iTerm.app.bak-999")
	if !found[bakApp] {
		t.Fatalf("missing bak app %q in %v", bakApp, got)
	}
}

func TestRunningITermApps_EnvWins(t *testing.T) {
	tmp := t.TempDir()
	envApp := filepath.Join(tmp, "custom", "iTerm.app")
	mkAppDir(t, envApp)
	homeApp := filepath.Join(tmp, "Applications", "iTerm.app")
	mkAppDir(t, homeApp)

	got := RunningITermAppsWithOpts(RunningITermAppOpts{
		Getenv: func(key string) string {
			if key == EnvITerm2AppPath {
				return envApp
			}
			return ""
		},
		Home: func() string { return tmp },
		IsApp: func(p string) bool {
			return p == envApp || p == homeApp
		},
		Pids: func() (string, error) { return "", nil },
	})

	if len(got) < 2 {
		t.Fatalf("got %v, want at least 2", got)
	}
	// Env path should be first
	if got[0] != envApp {
		t.Fatalf("first = %q, want env %q", got[0], envApp)
	}
}

func TestRunningITermApps_Dedup(t *testing.T) {
	tmp := t.TempDir()
	homeApp := filepath.Join(tmp, "Applications", "iTerm.app")

	// Same path appearing from disk + ps
	psOutput := homeApp + "/Contents/MacOS/iTerm2\n"

	got := RunningITermAppsWithOpts(RunningITermAppOpts{
		Getenv: func(string) string { return "" },
		Home:   func() string { return tmp },
		IsApp: func(p string) bool {
			return p == homeApp
		},
		Pids: func() (string, error) { return psOutput, nil },
	})

	count := 0
	for _, p := range got {
		if p == homeApp {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("home app appeared %d times, want 1; got %v", count, got)
	}
}

func TestRunningITermApps_NothingRunningOrOnDisk(t *testing.T) {
	got := RunningITermAppsWithOpts(RunningITermAppOpts{
		Getenv: func(string) string { return "" },
		Home:   func() string { return "" },
		IsApp:  func(string) bool { return false },
		Pids:   func() (string, error) { return "", nil },
	})

	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func mkAppDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(path, "Contents", "MacOS"), 0755); err != nil {
		t.Fatal(err)
	}
}