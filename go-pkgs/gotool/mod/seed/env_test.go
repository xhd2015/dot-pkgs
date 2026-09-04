package seed_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/seed"
)

func TestOverlayEnvSetsDirectAndInsteadOf(t *testing.T) {
	dir := gitInitTemp(t)
	env, err := seed.OverlayEnv([]string{
		"GOPROXY=https://proxy.golang.org,direct",
		"GOSUMDB=sum.golang.org",
		"GIT_CONFIG_COUNT=1",
		"KEEP=1",
	}, []seed.Mapping{{
		RepoDir:    dir,
		ModulePath: "github.com/xhd2015/lib/sub",
	}})
	if err != nil {
		t.Fatal(err)
	}
	byKey := envMap(env)
	if byKey["KEEP"] != "1" {
		t.Fatalf("KEEP lost: %v", env)
	}
	if byKey["GOPROXY"] != "direct" {
		t.Fatalf("GOPROXY=%q", byKey["GOPROXY"])
	}
	if byKey["GOSUMDB"] != "sum.golang.org" {
		t.Fatalf("GOSUMDB should be preserved, got %q", byKey["GOSUMDB"])
	}
	if byKey["GONOSUMDB"] != "github.com/xhd2015/lib/sub" {
		t.Fatalf("GONOSUMDB=%q", byKey["GONOSUMDB"])
	}
	if byKey["GOPRIVATE"] != "github.com/xhd2015/lib/sub" {
		t.Fatalf("GOPRIVATE=%q", byKey["GOPRIVATE"])
	}
	if byKey["GIT_CONFIG_NOSYSTEM"] != "1" {
		t.Fatal("missing GIT_CONFIG_NOSYSTEM")
	}
	if byKey["GIT_CONFIG_COUNT"] == "" || byKey["GIT_CONFIG_COUNT"] == "0" {
		t.Fatalf("GIT_CONFIG_COUNT=%q", byKey["GIT_CONFIG_COUNT"])
	}
	found := false
	for k, v := range byKey {
		if strings.HasPrefix(k, "GIT_CONFIG_VALUE_") && v == "https://github.com/xhd2015/lib" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing insteadOf https value: %v", env)
	}
	wantFile := "file://" + filepath.ToSlash(dir)
	foundKey := false
	for k, v := range byKey {
		if !strings.HasPrefix(k, "GIT_CONFIG_KEY_") {
			continue
		}
		if v == "url."+wantFile+".insteadOf" {
			foundKey = true
			break
		}
	}
	if !foundKey {
		t.Fatalf("missing insteadOf key for %s: %v", wantFile, env)
	}
}

func TestOverlayEnvNestedModuleDirUsesGitToplevel(t *testing.T) {
	root := gitInitTemp(t)
	sub := filepath.Join(root, "go-pkgs")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	env, err := seed.OverlayEnv(nil, []seed.Mapping{{
		RepoDir:    sub, // module dir, not git root
		ModulePath: "github.com/xhd2015/dot-pkgs/go-pkgs",
	}})
	if err != nil {
		t.Fatal(err)
	}
	wantFile := "file://" + filepath.ToSlash(root)
	found := false
	for k, v := range envMap(env) {
		if strings.HasPrefix(k, "GIT_CONFIG_KEY_") && v == "url."+wantFile+".insteadOf" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected insteadOf file URL %s, env=%v", wantFile, env)
	}
	// Must not point at the nested module dir.
	bad := "file://" + filepath.ToSlash(sub)
	for k, v := range envMap(env) {
		if strings.HasPrefix(k, "GIT_CONFIG_KEY_") && strings.Contains(v, bad) {
			t.Fatalf("insteadOf still uses module dir: %s", v)
		}
	}
}

func TestOverlayEnvSkipsUnknownHost(t *testing.T) {
	dir := t.TempDir() // no git required when host is not well-known
	env, err := seed.OverlayEnv([]string{
		"GOPROXY=file:///tmp/modproxy",
		"GOSUMDB=off",
		"KEEP=1",
	}, []seed.Mapping{{
		RepoDir:    dir,
		ModulePath: "example.com/lib",
	}})
	if err != nil {
		t.Fatal(err)
	}
	byKey := envMap(env)
	// No well-known mapping → preserve caller proxy (do not force direct).
	if byKey["GOPROXY"] != "file:///tmp/modproxy" {
		t.Fatalf("GOPROXY=%q", byKey["GOPROXY"])
	}
	if _, ok := byKey["GIT_CONFIG_COUNT"]; ok {
		t.Fatalf("unexpected GIT_CONFIG overlay: %v", env)
	}
	if byKey["KEEP"] != "1" {
		t.Fatalf("KEEP lost: %v", env)
	}
}

func TestOverlayEnvInvalidRepoDir(t *testing.T) {
	_, err := seed.OverlayEnv(nil, []seed.Mapping{{
		RepoDir:    filepath.Join(t.TempDir(), "missing"),
		ModulePath: "github.com/xhd2015/lib",
	}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOverlayEnvNotGitHardFail(t *testing.T) {
	dir := t.TempDir() // no git init
	_, err := seed.OverlayEnv(nil, []seed.Mapping{{
		RepoDir:    dir,
		ModulePath: "github.com/xhd2015/lib",
	}})
	if err == nil {
		t.Fatal("expected git toplevel error")
	}
	if !strings.Contains(err.Error(), "git toplevel") {
		t.Fatalf("err=%v", err)
	}
}

func TestOverlayEnvConflictSameVCS(t *testing.T) {
	a := gitInitTemp(t)
	b := gitInitTemp(t)
	_, err := seed.OverlayEnv(nil, []seed.Mapping{
		{RepoDir: a, ModulePath: "github.com/xhd2015/lib"},
		{RepoDir: b, ModulePath: "github.com/xhd2015/lib/sub"},
	})
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestOverlayEnvDedupesSameRepo(t *testing.T) {
	dir := gitInitTemp(t)
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	env, err := seed.OverlayEnv([]string{
		"GONOSUMDB=example.com/old",
		"GOPRIVATE=example.com/old",
	}, []seed.Mapping{
		{RepoDir: dir, ModulePath: "github.com/xhd2015/lib"},
		{RepoDir: sub, ModulePath: "github.com/xhd2015/lib/sub"},
	})
	if err != nil {
		t.Fatal(err)
	}
	byKey := envMap(env)
	// One VCS root → 6 insteadOf variants.
	if got := byKey["GIT_CONFIG_COUNT"]; got != "6" {
		t.Fatalf("GIT_CONFIG_COUNT=%q want 6", got)
	}
	if got := byKey["GONOSUMDB"]; got != "example.com/old,github.com/xhd2015/lib,github.com/xhd2015/lib/sub" {
		t.Fatalf("GONOSUMDB=%q", got)
	}
	if got := byKey["GOPRIVATE"]; got != "example.com/old,github.com/xhd2015/lib,github.com/xhd2015/lib/sub" {
		t.Fatalf("GOPRIVATE=%q", got)
	}
	if _, ok := byKey["GOSUMDB"]; ok {
		t.Fatalf("unexpected GOSUMDB force: %v", env)
	}
}

func envMap(env []string) map[string]string {
	out := make(map[string]string, len(env))
	for _, e := range env {
		k, v, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		out[k] = v
	}
	return out
}

func TestOverlayEnvNilBaseUsesEnviron(t *testing.T) {
	dir := gitInitTemp(t)
	t.Setenv("OVERLAY_ENV_MARK", "yes")
	env, err := seed.OverlayEnv(nil, []seed.Mapping{{
		RepoDir:    dir,
		ModulePath: "github.com/xhd2015/lib",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if envMap(env)["OVERLAY_ENV_MARK"] != "yes" {
		t.Fatalf("did not inherit environ: %v", env)
	}
}

func gitInitTemp(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	// Normalize for macOS /var → /private/var so file:// matches ShowToplevel.
	abs, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
