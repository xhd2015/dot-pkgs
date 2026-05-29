package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRejectsStagedSubModuleFile(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	smDir := filepath.Join(repo, "vendor", "libfoo")
	mustRun(t, repo, "mkdir", "-p", filepath.Join(smDir, "src"))
	writeFile(t, filepath.Join(smDir, "src", "main.c"), "int main() { return 0; }\n")
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)
	mustRun(t, repo, "git", "add", "vendor/libfoo/src/main.c")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if !errors.Is(err, errSubModuleFound) {
		t.Fatalf("expected submodule error, got %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "vendor/libfoo/") {
		t.Fatalf("expected vendor/libfoo/ in output, got:\n%s", got)
	}
}

func TestRunAllowsRegularStagedFile(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	writeFile(t, filepath.Join(repo, "hello.txt"), "hello world\n")
	mustRun(t, repo, "git", "add", "hello.txt")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if err != nil {
		t.Fatalf("expected no error for regular file, got %v\n%s", err, out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output, got:\n%s", out.String())
	}
}

func TestRunRejectsNestedFileInSubModule(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	smDir := filepath.Join(repo, "third_party", "dep")
	mustRun(t, repo, "mkdir", "-p", filepath.Join(smDir, "include"))
	writeFile(t, filepath.Join(smDir, "include", "dep.h"), "#pragma once\n")
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)
	mustRun(t, repo, "git", "add", "third_party/dep/include/dep.h")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if !errors.Is(err, errSubModuleFound) {
		t.Fatalf("expected submodule error for nested file, got %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "third_party/dep/") {
		t.Fatalf("expected third_party/dep/ in output, got:\n%s", got)
	}
}

func TestRunDeduplicatesSubModules(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	smDir := filepath.Join(repo, "vendor", "libbar")
	mustRun(t, repo, "mkdir", "-p", filepath.Join(smDir, "src"))
	writeFile(t, filepath.Join(smDir, "src", "foo.c"), "void foo() {}\n")
	writeFile(t, filepath.Join(smDir, "src", "bar.c"), "void bar() {}\n")
	writeFile(t, filepath.Join(smDir, "README.md"), "# libbar\n")
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)
	mustRun(t, repo, "git", "add", "vendor/libbar/src/foo.c", "vendor/libbar/src/bar.c", "vendor/libbar/README.md")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if !errors.Is(err, errSubModuleFound) {
		t.Fatalf("expected submodule error, got %v", err)
	}
	got := out.String()
	count := strings.Count(got, "vendor/libbar/")
	if count != 1 {
		t.Fatalf("expected vendor/libbar/ to appear once, got %d appearances:\n%s", count, got)
	}
}

func TestRunMixedSubModuleAndRegular(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	smDir := filepath.Join(repo, "vendor", "libbaz")
	mustRun(t, repo, "mkdir", "-p", filepath.Join(smDir, "src"))
	writeFile(t, filepath.Join(smDir, "src", "baz.c"), "void baz() {}\n")
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)
	writeFile(t, filepath.Join(repo, "hello.txt"), "hello world\n")
	mustRun(t, repo, "git", "add", "vendor/libbaz/src/baz.c", "hello.txt")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if !errors.Is(err, errSubModuleFound) {
		t.Fatalf("expected submodule error, got %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "vendor/libbaz/") {
		t.Fatalf("expected vendor/libbaz/ in output, got:\n%s", got)
	}
	if strings.Contains(got, "hello.txt") {
		t.Fatalf("did not expect hello.txt in output, got:\n%s", got)
	}
}

func TestRunNoStagedFiles(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if err != nil {
		t.Fatalf("expected no error with no staged files, got %v", err)
	}
}

func TestRunSubModuleWithGitFile(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	smDir := filepath.Join(repo, "vendor", "libgitfile")
	mustRun(t, repo, "mkdir", "-p", smDir)
	writeFile(t, filepath.Join(smDir, "code.go"), "package foo\n")
	writeFile(t, filepath.Join(smDir, ".git"), "gitdir: ../../.git/modules/vendor/libgitfile\n")
	mustRun(t, repo, "git", "add", "vendor/libgitfile/code.go")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if !errors.Is(err, errSubModuleFound) {
		t.Fatalf("expected submodule error for .git file, got %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "vendor/libgitfile/") {
		t.Fatalf("expected vendor/libgitfile/ in output, got:\n%s", got)
	}
}

func TestOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "git@git.xxx.com:team/repo.git")

	smDir := filepath.Join(repo, "vendor", "libq")
	mustRun(t, repo, "mkdir", "-p", smDir)
	writeFile(t, filepath.Join(smDir, "code.go"), "package q\n")
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)
	mustRun(t, repo, "git", "add", "vendor/libq/code.go")

	var out bytes.Buffer
	err := runWithOutput([]string{"--origin-domain", "other.example.com"}, &out)
	if err != nil {
		t.Fatalf("expected mismatched origin domain to skip, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain does not match, got:\n%s", out.String())
	}

	err = runWithOutput([]string{"--origin-domain", "git.xxx.com"}, &out)
	if !errors.Is(err, errSubModuleFound) {
		t.Fatalf("expected matching origin domain to scan, got %v", err)
	}
}

func TestExcludeOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "git@git.xxx.com:team/repo.git")

	smDir := filepath.Join(repo, "vendor", "libe")
	mustRun(t, repo, "mkdir", "-p", smDir)
	writeFile(t, filepath.Join(smDir, "code.go"), "package e\n")
	os.MkdirAll(filepath.Join(smDir, ".git"), 0755)
	mustRun(t, repo, "git", "add", "vendor/libe/code.go")

	var out bytes.Buffer
	err := runWithOutput([]string{"--exclude-origin-domain", "git.xxx.com"}, &out)
	if err != nil {
		t.Fatalf("expected matching excluded origin domain to skip, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain is excluded, got:\n%s", out.String())
	}

	err = runWithOutput([]string{"--exclude-origin-domain", "other.example.com"}, &out)
	if !errors.Is(err, errSubModuleFound) {
		t.Fatalf("expected non-excluded origin domain to scan, got %v", err)
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	mustRun(t, repo, "git", "init")
	mustRun(t, repo, "git", "config", "user.email", "test@example.com")
	mustRun(t, repo, "git", "config", "user.name", "Test User")
	return repo
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
}
