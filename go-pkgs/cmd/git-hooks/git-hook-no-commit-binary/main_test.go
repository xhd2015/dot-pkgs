package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/detect"
)

func TestRunRejectsStagedBinary(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	binaryPath := filepath.Join(repo, "program.bin")
	writeBinaryFile(t, binaryPath, 1024)
	mustRun(t, repo, "git", "add", "program.bin")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if !errors.Is(err, errBinaryFilesFound) {
		t.Fatalf("expected binary error, got %v", err)
	}
	if !strings.Contains(out.String(), "program.bin") {
		t.Fatalf("expected program.bin in output, got:\n%s", out.String())
	}
}

func TestRunAllowsStagedTextFile(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	writeFile(t, filepath.Join(repo, "hello.txt"), "hello world\n")
	mustRun(t, repo, "git", "add", "hello.txt")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if err != nil {
		t.Fatalf("expected no error for text file, got %v\n%s", err, out.String())
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output, got:\n%s", out.String())
	}
}

func TestRunAllowsEmptyFile(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	writeFile(t, filepath.Join(repo, "empty.txt"), "")
	mustRun(t, repo, "git", "add", "empty.txt")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if err != nil {
		t.Fatalf("expected no error for empty file, got %v\n%s", err, out.String())
	}
}

func TestRunMixedBinaryAndText(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	writeFile(t, filepath.Join(repo, "hello.txt"), "hello world\n")
	writeBinaryFile(t, filepath.Join(repo, "bad.bin"), 512)
	mustRun(t, repo, "git", "add", "hello.txt", "bad.bin")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if !errors.Is(err, errBinaryFilesFound) {
		t.Fatalf("expected binary error, got %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "bad.bin") {
		t.Fatalf("expected bad.bin in output, got:\n%s", got)
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

func TestRunAllowsStagedSubmodule(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	subRepo := initGitRepo(t)
	writeFile(t, filepath.Join(subRepo, "subfile.txt"), "content\n")
	mustRun(t, subRepo, "git", "add", "subfile.txt")
	mustRun(t, subRepo, "git", "commit", "-m", "init submodule")

	mustRun(t, repo, "git", "-c", "protocol.file.allow=always", "submodule", "add", subRepo, "x")
	mustRun(t, repo, "git", "add", "x")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if err != nil {
		t.Fatalf("expected no error for staged submodule, got %v\n%s", err, out.String())
	}
}

func TestRunSkipsDirectoryEntries(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)

	subRepo := initGitRepo(t)
	writeFile(t, filepath.Join(subRepo, "hello.txt"), "hello\n")
	mustRun(t, subRepo, "git", "add", "hello.txt")
	mustRun(t, subRepo, "git", "commit", "-m", "init")

	mustRun(t, repo, "git", "-c", "protocol.file.allow=always", "submodule", "add", subRepo, "dot-pkgs")
	mustRun(t, repo, "git", "add", "dot-pkgs")

	var out bytes.Buffer
	err := runWithOutput(nil, &out)
	if err != nil {
		t.Fatalf("expected no error when staged entry is a directory (submodule), got %v\n%s", err, out.String())
	}
}

func TestOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "git@git.xxx.com:team/repo.git")
	writeBinaryFile(t, filepath.Join(repo, "bad.bin"), 512)
	mustRun(t, repo, "git", "add", "bad.bin")

	var out bytes.Buffer
	err := runWithOutput([]string{"--origin-domain", "other.example.com"}, &out)
	if err != nil {
		t.Fatalf("expected mismatched origin domain to skip, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain does not match, got:\n%s", out.String())
	}

	err = runWithOutput([]string{"--origin-domain", "git.xxx.com"}, &out)
	if !errors.Is(err, errBinaryFilesFound) {
		t.Fatalf("expected matching origin domain to scan, got %v", err)
	}
}

func TestExcludeOriginDomainGate(t *testing.T) {
	repo := initGitRepo(t)
	t.Chdir(repo)
	mustRun(t, repo, "git", "remote", "add", "origin", "git@git.xxx.com:team/repo.git")
	writeBinaryFile(t, filepath.Join(repo, "bad.bin"), 512)
	mustRun(t, repo, "git", "add", "bad.bin")

	var out bytes.Buffer
	err := runWithOutput([]string{"--exclude-origin-domain", "git.xxx.com"}, &out)
	if err != nil {
		t.Fatalf("expected matching excluded origin domain to skip, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output when origin domain is excluded, got:\n%s", out.String())
	}

	err = runWithOutput([]string{"--exclude-origin-domain", "other.example.com"}, &out)
	if !errors.Is(err, errBinaryFilesFound) {
		t.Fatalf("expected non-excluded origin domain to scan, got %v", err)
	}
}

func TestDetectFileType(t *testing.T) {
	dir := t.TempDir()

	t.Run("empty file", func(t *testing.T) {
		writeFile(t, filepath.Join(dir, "empty.txt"), "")
		_, isBin, err := detect.DetectFileType(filepath.Join(dir, "empty.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if isBin {
			t.Fatal("empty file should not be binary")
		}
	})

	t.Run("text file", func(t *testing.T) {
		writeFile(t, filepath.Join(dir, "hello.txt"), "hello world\n")
		_, isBin, err := detect.DetectFileType(filepath.Join(dir, "hello.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if isBin {
			t.Fatal("text file should not be binary")
		}
	})

	t.Run("binary file", func(t *testing.T) {
		writeBinaryFile(t, filepath.Join(dir, "bad.bin"), 512)
		desc, isBin, err := detect.DetectFileType(filepath.Join(dir, "bad.bin"))
		if err != nil {
			t.Fatal(err)
		}
		if !isBin {
			t.Fatal("binary file should be detected as binary")
		}
		if desc == "" {
			t.Fatal("description should not be empty for binary file")
		}
	})

	t.Run("png image", func(t *testing.T) {
		pngData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
			0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
			0x54, 0x08, 0xD7, 0x63, 0x60, 0x60, 0x60, 0x00,
			0x00, 0x00, 0x04, 0x00, 0x01, 0x23, 0x45, 0x67,
			0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44,
			0xAE, 0x42, 0x60, 0x82,
		}
		pngPath := filepath.Join(dir, "test.png")
		if err := os.WriteFile(pngPath, pngData, 0644); err != nil {
			t.Fatal(err)
		}
		_, isBin, err := detect.DetectFileType(pngPath)
		if err != nil {
			t.Fatal(err)
		}
		if !isBin {
			t.Fatal("PNG file should be detected as binary")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, _, err := detect.DetectFileType(filepath.Join(dir, "nonexistent"))
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})

	t.Run("utf8 with null bytes", func(t *testing.T) {
		data := append([]byte("hello"), 0x00, 0x01, 0x02)
		path := filepath.Join(dir, "nulls.bin")
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		_, isBin, err := detect.DetectFileType(path)
		if err != nil {
			t.Fatal(err)
		}
		if !isBin {
			t.Fatal("file with null bytes should be binary")
		}
	})
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeBinaryFile(t *testing.T, path string, size int) {
	t.Helper()
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
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
