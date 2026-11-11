package service

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/lookpath"
)

func baseEnv(dir string) []string {
	return []string{"HOME=" + dir, "PATH=/usr/bin:/bin"}
}

func TestBuildEnvLayersExtraEnvAndExtraPath(t *testing.T) {
	extra := t.TempDir()
	env, err := BuildEnv(EnvOptions{
		Base:        baseEnv(extra),
		Extra:       map[string]string{"FOO": "bar", "PATH": "/usr/bin:/bin"},
		ExtraPath:   extra,
		LookupTools: []string{},
	})
	if err != nil {
		t.Fatalf("BuildEnv() error = %v", err)
	}
	if got := EnvValue(env, "FOO"); got != "bar" {
		t.Fatalf("FOO = %q", got)
	}
	if got := EnvValue(env, "PATH"); got != extra+string(os.PathListSeparator)+"/usr/bin:/bin" {
		t.Fatalf("PATH = %q, want extra dir prepended", got)
	}
}

func TestBuildEnvDedupesExtraPathAgainstExistingPath(t *testing.T) {
	dir := t.TempDir()
	env, err := BuildEnv(EnvOptions{
		Base:        []string{"PATH=/usr/bin:" + dir},
		ExtraPath:   dir + string(os.PathListSeparator) + dir + string(os.PathListSeparator) + "/usr/bin",
		LookupTools: []string{},
	})
	if err != nil {
		t.Fatalf("BuildEnv() error = %v", err)
	}
	if got, want := EnvValue(env, "PATH"), "/usr/bin:"+dir; got != want {
		t.Fatalf("PATH = %q, want %q", got, want)
	}
}

func TestBuildEnvIgnoresInvalidExtraPath(t *testing.T) {
	env, err := BuildEnv(EnvOptions{
		Base:        baseEnv(t.TempDir()),
		ExtraPath:   "relative/bin",
		LookupTools: []string{},
	})
	if err != nil {
		t.Fatalf("BuildEnv() error = %v, want the invalid layer skipped", err)
	}
	if got := EnvValue(env, "PATH"); got != "/usr/bin:/bin" {
		t.Fatalf("PATH = %q, want the base path", got)
	}
}

func TestBuildEnvMergesLoginShellEnv(t *testing.T) {
	env, err := BuildEnv(EnvOptions{
		Base:            baseEnv(t.TempDir()),
		InheritShellEnv: true,
		LoginEnv: &lookpath.LoginEnvOptions{
			DetectShell: func() string { return "bash" },
			RunLogin: func(shell, command string, env []string) (string, error) {
				return "SHELL_ONLY=1\x00PATH=/login/bin", nil
			},
		},
		LookupTools: []string{},
	})
	if err != nil {
		t.Fatalf("BuildEnv() error = %v", err)
	}
	if got := EnvValue(env, "SHELL_ONLY"); got != "1" {
		t.Fatalf("SHELL_ONLY = %q, want the login env layer merged", got)
	}
	if got := EnvValue(env, "PATH"); got != "/login/bin" {
		t.Fatalf("PATH = %q, want the login shell path", got)
	}
}

func TestBuildEnvReturnsShellProbeErrorButKeepsEnv(t *testing.T) {
	probeErr := errors.New("shell probe failed")
	env, err := BuildEnv(EnvOptions{
		Base:            baseEnv(t.TempDir()),
		InheritShellEnv: true,
		LoginEnv: &lookpath.LoginEnvOptions{
			DetectShell: func() string { return "bash" },
			RunLogin: func(shell, command string, env []string) (string, error) {
				return "", probeErr
			},
		},
		LookupTools: []string{},
	})
	if !errors.Is(err, probeErr) {
		t.Fatalf("BuildEnv() error = %v, want %v", err, probeErr)
	}
	if got := EnvValue(env, "PATH"); got != "/usr/bin:/bin" {
		t.Fatalf("PATH = %q, want the base env to stay usable", got)
	}
}

func TestBuildEnvRejectsEmptyLoginEnv(t *testing.T) {
	_, err := BuildEnv(EnvOptions{
		Base:            baseEnv(t.TempDir()),
		InheritShellEnv: true,
		LoginEnv: &lookpath.LoginEnvOptions{
			DetectShell: func() string { return "bash" },
			RunLogin: func(shell, command string, env []string) (string, error) {
				return "", nil
			},
		},
		LookupTools: []string{},
	})
	if err == nil || !strings.Contains(err.Error(), "empty environment") {
		t.Fatalf("BuildEnv() error = %v, want empty environment rejection", err)
	}
}

func TestBuildEnvUsesProcessEnvWhenBaseIsNil(t *testing.T) {
	t.Setenv("SERVICE_TEST_MARKER", "present")
	env, err := BuildEnv(EnvOptions{LookupTools: []string{}})
	if err != nil {
		t.Fatalf("BuildEnv() error = %v", err)
	}
	if got := EnvValue(env, "SERVICE_TEST_MARKER"); got != "present" {
		t.Fatalf("SERVICE_TEST_MARKER = %q, want the process env inherited", got)
	}
}

func TestNormalizeExtraPath(t *testing.T) {
	dir := t.TempDir()
	got, err := NormalizeExtraPath("  " + dir + string(os.PathListSeparator) + dir + "  ")
	if err != nil {
		t.Fatalf("NormalizeExtraPath() error = %v", err)
	}
	if got != dir {
		t.Fatalf("NormalizeExtraPath() = %q, want %q", got, dir)
	}
	if got, err := NormalizeExtraPath(""); err != nil || got != "" {
		t.Fatalf("NormalizeExtraPath(\"\") = %q, %v", got, err)
	}
}

func TestNormalizeExtraPathRejections(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantSub string
	}{
		{"empty entry", "/usr/bin:" + string(os.PathListSeparator) + "/bin", "empty entry"},
		{"variable", "$HOME/bin", "does not expand variables"},
		{"backtick", "`pwd`", "does not expand variables"},
		{"root", "/", "invalid extra path directory"},
		{"relative", "bin/tools", "must be absolute"},
		{"newline", "/usr/bin\n/bin", "single line"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NormalizeExtraPath(tc.value); err == nil || !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("NormalizeExtraPath(%q) error = %v, want %q", tc.value, err, tc.wantSub)
			}
		})
	}
}

func TestNormalizeExtraPathExpandsHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("no home directory")
	}
	got, err := NormalizeExtraPath("~/bin")
	if err != nil {
		t.Fatalf("NormalizeExtraPath() error = %v", err)
	}
	if got != filepath.Join(home, "bin") {
		t.Fatalf("NormalizeExtraPath() = %q, want %q", got, filepath.Join(home, "bin"))
	}
}
