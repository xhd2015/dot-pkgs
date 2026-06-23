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

func TestBuildScriptCreatesWindowAndCDs(t *testing.T) {
	script := BuildScript("/tmp/proj")
	if !strings.Contains(script, "create window with default profile") {
		t.Fatalf("script must create a window: %q", script)
	}
	if strings.Contains(script, "create tab") {
		t.Fatal("script must not create a tab")
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

func TestOpenConfigRejectsNonDirectory(t *testing.T) {
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
	err := OpenConfig(t.TempDir(), &Config{
		Installed: func() bool { return false },
	})
	if !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("OpenConfig() error = %v, want ErrNotInstalled", err)
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