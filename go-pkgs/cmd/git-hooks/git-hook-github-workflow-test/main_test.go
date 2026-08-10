package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	cfg, err := parseArgs([]string{"--fix", "--origin-domain=github.com", "--exclude-origin-domain", "git.example.com"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.fix {
		t.Fatal("expected --fix to be enabled")
	}
	if cfg.domainFilter.OriginDomain != "github.com" {
		t.Fatalf("origin domain = %q, want github.com", cfg.domainFilter.OriginDomain)
	}
	if cfg.domainFilter.ExcludeOriginDomain != "git.example.com" {
		t.Fatalf("exclude origin domain = %q, want git.example.com", cfg.domainFilter.ExcludeOriginDomain)
	}
}

func TestParseArgsUnknownFlag(t *testing.T) {
	_, err := parseArgs([]string{"--unknown"}, io.Discard)
	if err == nil {
		t.Fatal("expected unknown flag error")
	}
	if !strings.Contains(err.Error(), "unknown flag: --unknown") {
		t.Fatalf("expected unknown flag error, got %v", err)
	}
}

func TestWorkflowContent(t *testing.T) {
	content, err := workflowContent([]goModule{
		{Dir: "", GoVersion: "1.22"},
		{Dir: "sub-nested-dir", GoVersion: "1.23.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"container:",
		"golang:1.23.0",
		"go test -v ./...",
		"go -C sub-nested-dir test -v ./...",
		"command -v doctest",
		"go install github.com/xhd2015/doctest/cmd/doctest@latest",
		"doctest test -v --label-all ./...",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected workflow to contain %q, got:\n%s", want, content)
		}
	}
	goTest := strings.Index(content, "go test -v ./...")
	install := strings.Index(content, "go install github.com/xhd2015/doctest/cmd/doctest@latest")
	doctest := strings.Index(content, "doctest test -v --label-all ./...")
	if goTest < 0 || install < 0 || doctest < 0 || goTest > install || install > doctest {
		t.Fatalf("expected go test before doctest install before doctest run, got:\n%s", content)
	}
}

func TestWorkflowContentWithoutGoModules(t *testing.T) {
	content, err := workflowContent(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"golang:latest",
		"no go.mod files found; skipping go test",
		"doctest test -v --label-all ./...",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected workflow to contain %q, got:\n%s", want, content)
		}
	}
}

func TestGoModVersion(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("go.mod", []byte("module example.com/repo\n\ngo 1.23.4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	version, err := goModVersion("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if version != "1.23.4" {
		t.Fatalf("version = %q, want 1.23.4", version)
	}
}

func TestDiscoverGoModules(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "git", "init")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub-nested-dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub-nested-dir", "go.mod"), []byte("module example.com/sub\n\ngo 1.23.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "fixture", "testdata", "module"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fixture", "testdata", "module", "go.mod"), []byte("module example.com/fixture\n\ngo 1.99\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("ignored-dir/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "ignored-dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored-dir", "go.mod"), []byte("module example.com/ignored\n\ngo 1.99\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	modules, err := discoverGoModules(dir, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) != 2 {
		t.Fatalf("module count = %d, want 2: %#v", len(modules), modules)
	}
	if modules[0].Dir != "" || modules[0].GoVersion != "1.22" {
		t.Fatalf("root module = %#v", modules[0])
	}
	if modules[1].Dir != "sub-nested-dir" || modules[1].GoVersion != "1.23.0" {
		t.Fatalf("nested module = %#v", modules[1])
	}
}

func TestDiscoverGoModulesSkipsMissingGoDirective(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "git", "init")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	invalidDir := filepath.Join(dir, "kool-template")
	if err := os.MkdirAll(invalidDir, 0o755); err != nil {
		t.Fatal(err)
	}
	invalidGoMod := filepath.Join(invalidDir, "go.mod")
	if err := os.WriteFile(invalidGoMod, []byte("// Not a Go module.\n// Template only.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var warn strings.Builder
	modules, err := discoverGoModules(dir, &warn)
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) != 1 {
		t.Fatalf("module count = %d, want 1: %#v", len(modules), modules)
	}
	if modules[0].Dir != "" || modules[0].GoVersion != "1.22" {
		t.Fatalf("root module = %#v", modules[0])
	}
	if !strings.Contains(warn.String(), "warning:") || !strings.Contains(warn.String(), "missing go directive") {
		t.Fatalf("expected missing go directive warning, got:\n%s", warn.String())
	}
	if !strings.Contains(warn.String(), invalidGoMod) {
		t.Fatalf("expected warning to mention %s, got:\n%s", invalidGoMod, warn.String())
	}
}

func TestGoModVersionMissingGoDirective(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(path, []byte("// Not a Go module.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := goModVersion(path)
	if err == nil {
		t.Fatal("expected missing go directive error")
	}
	if !errors.Is(err, errMissingGoDirective) {
		t.Fatalf("err = %v, want errors.Is(..., errMissingGoDirective)", err)
	}
}

func TestRunFixFromSubdirCreatesWorkflowAtRepoRoot(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "git", "init")
	mustRun(t, dir, "git", "remote", "add", "origin", "git@github.com:owner/repo.git")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/repo\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	subdir := filepath.Join(dir, "task-hub")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(subdir)
	var out strings.Builder
	if err := runWithOutput([]string{"--fix"}, &out); err != nil {
		t.Fatal(err)
	}
	rootWorkflow := filepath.Join(dir, ".github", "workflows", "test.yml")
	if _, err := os.Stat(rootWorkflow); err != nil {
		t.Fatalf("expected workflow at repo root %s: %v", rootWorkflow, err)
	}
	subdirWorkflow := filepath.Join(subdir, ".github", "workflows", "test.yml")
	if _, err := os.Stat(subdirWorkflow); !os.IsNotExist(err) {
		t.Fatalf("workflow must not be created in subdirectory %s", subdirWorkflow)
	}
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}
