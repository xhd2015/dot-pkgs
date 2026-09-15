package install

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBinNameFromPackage(t *testing.T) {
	cases := []struct {
		pkg, mod, want string
	}{
		{"./cmd/browser-agent", "github.com/xhd2015/browser-agent", "browser-agent"},
		{"cmd/browser-agent", "", "browser-agent"},
		{".", "github.com/xhd2015/browser-agent", "browser-agent"},
		{"./", "github.com/foo/mytool", "mytool"},
		{"github.com/foo/bar/cmd/x", "", "x"},
		{"./cmd/go-best-practice", "github.com/xhd2015/go-best-practice", "go-best-practice"},
	}
	for _, tc := range cases {
		got := BinNameFromPackage(tc.pkg, tc.mod)
		if got != tc.want {
			t.Errorf("BinNameFromPackage(%q, %q)=%q want %q", tc.pkg, tc.mod, got, tc.want)
		}
	}
}

func TestResolvePrimary_LookPathWins(t *testing.T) {
	got, fromPATH, err := ResolvePrimary("mytool", "/home/u", resolveHooks{
		LookPath: func(string) (string, error) {
			return "/opt/bin/mytool", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !fromPATH || got != "/opt/bin/mytool" {
		t.Fatalf("got %q fromPATH=%v", got, fromPATH)
	}
}

func TestResolvePrimary_DefaultLocalBin(t *testing.T) {
	got, fromPATH, err := ResolvePrimary("mytool", "/home/u", resolveHooks{
		LookPath: func(string) (string, error) {
			return "", fmt.Errorf("not found")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/home/u", ".local", "bin", "mytool")
	if fromPATH || got != want {
		t.Fatalf("got %q fromPATH=%v want %q", got, fromPATH, want)
	}
}

func TestExtraTargets_ExistingCopies(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	local := filepath.Join(home, ".local", "bin", "tool")
	gopathBin := filepath.Join(dir, "gopath", "bin", "tool")
	gobin := filepath.Join(dir, "gobin", "tool")
	primary := filepath.Join(dir, "path", "tool")
	for _, p := range []string{local, gopathBin, gobin, primary} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	extras := ExtraTargets("tool", home, primary, func(name string) (string, error) {
		switch name {
		case "GOBIN":
			return filepath.Join(dir, "gobin"), nil
		case "GOPATH":
			return filepath.Join(dir, "gopath"), nil
		default:
			return "", nil
		}
	})
	if len(extras) != 3 {
		t.Fatalf("extras=%v", extras)
	}
	for _, e := range extras {
		if sameFilePath(e, primary) {
			t.Fatalf("primary leaked into extras: %v", extras)
		}
	}
}

func TestExtraTargets_SkipsMissing(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	_ = os.MkdirAll(filepath.Join(home, ".local", "bin"), 0o755)
	primary := filepath.Join(dir, "path", "tool")
	_ = os.MkdirAll(filepath.Dir(primary), 0o755)
	_ = os.WriteFile(primary, []byte("x"), 0o755)
	extras := ExtraTargets("tool", home, primary, func(string) (string, error) {
		return filepath.Join(dir, "gopath"), nil
	})
	if len(extras) != 0 {
		t.Fatalf("want no extras, got %v", extras)
	}
}

func TestInstall_FreshToLocalBin(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "mod")
	_ = os.MkdirAll(modDir, 0o755)
	_ = os.WriteFile(filepath.Join(modDir, "go.mod"), []byte("module example.com/mytool\n"), 0o644)
	home := filepath.Join(dir, "home")

	var ensuredDir string
	var stdout bytes.Buffer
	res, err := Install(Options{
		Dir:     modDir,
		Package: ".",
		LookPath: func(string) (string, error) {
			return "", fmt.Errorf("missing")
		},
		UserHome: func() (string, error) { return home, nil },
		GOOS:     "linux",
		Build: func(_, _, out string) error {
			if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				return err
			}
			return os.WriteFile(out, []byte("new"), 0o755)
		},
		EnsurePATH: func(destDir string, _ io.Writer) error {
			ensuredDir = destDir
			return nil
		},
		Stdout: &stdout,
		Stderr: io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".local", "bin", "mytool")
	if res.Primary != want {
		t.Fatalf("primary=%q want %q", res.Primary, want)
	}
	if len(res.Extras) != 0 {
		t.Fatalf("extras=%v", res.Extras)
	}
	if !res.PATHEnsured || ensuredDir != filepath.Join(home, ".local", "bin") {
		t.Fatalf("PATHEnsured=%v ensuredDir=%q", res.PATHEnsured, ensuredDir)
	}
	if !strings.Contains(stdout.String(), "default ~/.local/bin") {
		t.Fatalf("stdout:\n%s", stdout.String())
	}
}

func TestInstall_LookPathPlusGOPATHExtra(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "mod")
	_ = os.MkdirAll(modDir, 0o755)
	_ = os.WriteFile(filepath.Join(modDir, "go.mod"), []byte("module example.com/app\n"), 0o644)
	home := filepath.Join(dir, "home")
	primary := filepath.Join(dir, "path", "tool")
	gopathBin := filepath.Join(dir, "gopath", "bin", "tool")
	for _, p := range []string{primary, gopathBin} {
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte("old"), 0o755)
	}

	var signed []string
	res, err := Install(Options{
		Dir:     modDir,
		Package: "./cmd/tool",
		LookPath: func(string) (string, error) {
			return primary, nil
		},
		GoEnv: func(name string) (string, error) {
			if name == "GOPATH" {
				return filepath.Join(dir, "gopath"), nil
			}
			return "", nil
		},
		UserHome: func() (string, error) { return home, nil },
		GOOS:     "darwin",
		Build: func(_, _, out string) error {
			return os.WriteFile(out, []byte("fresh"), 0o755)
		},
		CodeSign: func(path string) error {
			signed = append(signed, path)
			return nil
		},
		EnsurePATH: func(string, io.Writer) error { return nil },
		Stdout:     io.Discard,
		Stderr:     io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Primary != primary {
		t.Fatalf("primary=%q", res.Primary)
	}
	if len(res.Extras) != 1 || res.Extras[0] != gopathBin {
		t.Fatalf("extras=%v", res.Extras)
	}
	data, _ := os.ReadFile(gopathBin)
	if string(data) != "fresh" {
		t.Fatalf("gopath content=%q", data)
	}
	if len(signed) != 2 {
		t.Fatalf("signed=%v", signed)
	}
	// LookPath primary is not under ~/.local/bin → EnsurePATH should not run
	if res.PATHEnsured {
		t.Fatal("PATHEnsured should be false when not writing to ~/.local/bin")
	}
}

func TestInstall_CodesignWarningContinues(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "mod")
	_ = os.MkdirAll(modDir, 0o755)
	_ = os.WriteFile(filepath.Join(modDir, "go.mod"), []byte("module example.com/x\n"), 0o644)
	home := filepath.Join(dir, "home")
	primary := filepath.Join(dir, "bin", "x")
	_ = os.MkdirAll(filepath.Dir(primary), 0o755)
	_ = os.WriteFile(primary, []byte("o"), 0o755)

	var stderr bytes.Buffer
	res, err := Install(Options{
		Dir:        modDir,
		Package:    ".",
		LookPath:   func(string) (string, error) { return primary, nil },
		UserHome:   func() (string, error) { return home, nil },
		GOOS:       "darwin",
		Build:      func(_, _, out string) error { return os.WriteFile(out, []byte("n"), 0o755) },
		CodeSign:   func(string) error { return fmt.Errorf("denied") },
		EnsurePATH: func(string, io.Writer) error { return nil },
		Stdout:     io.Discard,
		Stderr:     &stderr,
	})
	if err != nil {
		t.Fatalf("install must succeed: %v", err)
	}
	if len(res.Signed) != 0 {
		t.Fatalf("signed=%v", res.Signed)
	}
	if !strings.Contains(stderr.String(), "warning: codesign") {
		t.Fatalf("stderr=%s", stderr.String())
	}
}

func TestInstall_StagingDirSkipsPATHAndExtras(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "mod")
	_ = os.MkdirAll(modDir, 0o755)
	_ = os.WriteFile(filepath.Join(modDir, "go.mod"), []byte("module example.com/mytool\n"), 0o644)
	stage := filepath.Join(dir, "stage")
	home := filepath.Join(dir, "home")
	pathHit := filepath.Join(dir, "path", "mytool")
	_ = os.MkdirAll(filepath.Dir(pathHit), 0o755)
	_ = os.WriteFile(pathHit, []byte("old"), 0o755)

	var ensured bool
	res, err := Install(Options{
		Dir:          modDir,
		Package:      ".",
		InstallToDir: stage,
		LookPath:     func(string) (string, error) { return pathHit, nil },
		UserHome:     func() (string, error) { return home, nil },
		Build: func(_, _, out string) error {
			if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				return err
			}
			return os.WriteFile(out, []byte("staged"), 0o755)
		},
		EnsurePATH: func(string, io.Writer) error {
			ensured = true
			return nil
		},
		Stdout: io.Discard,
		Stderr: io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(stage, "mytool")
	if res.Primary != want {
		t.Fatalf("primary=%q want %q", res.Primary, want)
	}
	if len(res.Extras) != 0 {
		t.Fatalf("extras=%v", res.Extras)
	}
	if res.PATHEnsured || ensured {
		t.Fatal("staging must not EnsurePATH")
	}
	data, _ := os.ReadFile(pathHit)
	if string(data) != "old" {
		t.Fatalf("LookPath copy mutated: %q", data)
	}
}

func TestTargetGOOSArch_PrefersInstallEnv(t *testing.T) {
	t.Setenv(EnvInstallGOOS, "linux")
	t.Setenv(EnvInstallGOARCH, "arm64")
	t.Setenv("GOOS", "darwin")
	t.Setenv("GOARCH", "amd64")
	if TargetGOOS() != "linux" || TargetGOARCH() != "arm64" {
		t.Fatalf("got %s/%s", TargetGOOS(), TargetGOARCH())
	}
}

func TestProductBuildEnv_CrossStripsGOFLAGS(t *testing.T) {
	t.Setenv(EnvInstallGOOS, "plan9")
	t.Setenv(EnvInstallGOARCH, "amd64")
	env := ProductBuildEnv([]string{"GOFLAGS=-linkmode=external", "PATH=/bin", "GOOS=darwin"})
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "GOFLAGS=") {
		t.Fatalf("GOFLAGS should be stripped: %v", env)
	}
	if !strings.Contains(joined, "GOOS=plan9") || !strings.Contains(joined, "GOARCH=amd64") {
		t.Fatalf("target missing: %v", env)
	}
	if !strings.Contains(joined, "CGO_ENABLED=0") {
		t.Fatalf("want CGO_ENABLED=0: %v", env)
	}
}

func TestInstall_StagingDirFromEnv(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "mod")
	_ = os.MkdirAll(modDir, 0o755)
	_ = os.WriteFile(filepath.Join(modDir, "go.mod"), []byte("module example.com/x\n"), 0o644)
	stage := filepath.Join(dir, "envstage")
	t.Setenv(EnvInstallToDir, stage)
	res, err := Install(Options{
		Dir:     modDir,
		Package: ".",
		LookPath: func(string) (string, error) {
			return "", fmt.Errorf("missing")
		},
		UserHome: func() (string, error) { return filepath.Join(dir, "home"), nil },
		Build: func(_, _, out string) error {
			if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				return err
			}
			return os.WriteFile(out, []byte("e"), 0o755)
		},
		EnsurePATH: func(string, io.Writer) error { return nil },
		Stdout:     io.Discard,
		Stderr:     io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Primary != filepath.Join(stage, "x") {
		t.Fatalf("primary=%q", res.Primary)
	}
}
