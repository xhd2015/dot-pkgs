package tidy_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/seed"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/tidy"
)

func TestSeededResolvesLocalTagOffline(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	scene := t.TempDir()
	modCache := filepath.Join(scene, "gomodcache")
	t.Cleanup(func() { chmodTreeWritable(modCache) })

	lib := filepath.Join(scene, "lib")
	mustMkdir(t, lib)
	mustWrite(t, filepath.Join(lib, "go.mod"), "module github.com/xhd2015/wrk-seeded-tidy\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(lib, "hello.go"), "package seededtidy\n\nfunc Hello() string { return \"hi\" }\n")
	git(t, lib, "init", "-b", "main")
	git(t, lib, "config", "user.name", "expt")
	git(t, lib, "config", "user.email", "expt@example.com")
	git(t, lib, "add", ".")
	git(t, lib, "commit", "-m", "init")
	git(t, lib, "tag", "v0.0.1")

	consume := filepath.Join(scene, "consume")
	mustMkdir(t, consume)
	mustWrite(t, filepath.Join(consume, "go.mod"), ""+
		"module consume\n\n"+
		"go 1.22\n\n"+
		"require github.com/xhd2015/wrk-seeded-tidy v0.0.1\n")
	mustWrite(t, filepath.Join(consume, "main.go"), ""+
		"package main\n\n"+
		"import (\n"+
		"\t\"fmt\"\n"+
		"\tseededtidy \"github.com/xhd2015/wrk-seeded-tidy\"\n"+
		")\n\n"+
		"func main() { fmt.Print(seededtidy.Hello()) }\n")

	baseEnv := append(os.Environ(),
		"GOMODCACHE="+modCache,
		"GOCACHE="+filepath.Join(scene, "gocache"),
		"GOPATH="+filepath.Join(scene, "gopath"),
		"GOTOOLCHAIN=local",
		"https_proxy=http://127.0.0.1:1",
		"HTTPS_PROXY=http://127.0.0.1:1",
		"http_proxy=http://127.0.0.1:1",
		"HTTP_PROXY=http://127.0.0.1:1",
	)

	if err := tidy.Seeded(context.Background(), tidy.SeededRequest{
		Dir: consume,
		Locals: []seed.Mapping{{
			RepoDir:    lib,
			ModulePath: "github.com/xhd2015/wrk-seeded-tidy",
		}},
		Environ: baseEnv,
	}); err != nil {
		t.Fatalf("Seeded: %v", err)
	}

	sum, err := os.ReadFile(filepath.Join(consume, "go.sum"))
	if err != nil {
		t.Fatalf("go.sum: %v", err)
	}
	if !strings.Contains(string(sum), "github.com/xhd2015/wrk-seeded-tidy v0.0.1") {
		t.Fatalf("go.sum missing module:\n%s", sum)
	}
	extracted := filepath.Join(modCache, "github.com/xhd2015/wrk-seeded-tidy@v0.0.1", "hello.go")
	if _, err := os.Stat(extracted); err != nil {
		t.Fatalf("module not in GOMODCACHE: %v", err)
	}

	// Offline build with no overlay (cache already filled by tidy).
	cmd := exec.Command("go", "build", "-o", filepath.Join(scene, "out"), ".")
	cmd.Dir = consume
	cmd.Env = append(append([]string{}, baseEnv...), "GOPROXY=off", "GOSUMDB=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
}

func TestSeededEmptyLocalsPlainTidy(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "go.mod"), "module plain\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n")
	if err := tidy.Seeded(context.Background(), tidy.SeededRequest{Dir: dir}); err != nil {
		t.Fatal(err)
	}
}

func TestSeededInvalidLocalRepo(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "go.mod"), "module plain\n\ngo 1.22\n")
	err := tidy.Seeded(context.Background(), tidy.SeededRequest{
		Dir: dir,
		Locals: []seed.Mapping{{
			RepoDir:    filepath.Join(dir, "missing-repo"),
			ModulePath: "github.com/xhd2015/lib",
		}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSeededNestedModuleDirAsRepoDir(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}

	scene := t.TempDir()
	modCache := filepath.Join(scene, "gomodcache")
	t.Cleanup(func() { chmodTreeWritable(modCache) })

	root := filepath.Join(scene, "dot-pkgs")
	mustMkdir(t, root)
	mod := filepath.Join(root, "go-pkgs")
	mustMkdir(t, mod)
	mustWrite(t, filepath.Join(mod, "go.mod"), "module github.com/xhd2015/wrk-seeded-nested/go-pkgs\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(mod, "hello.go"), "package gopkgs\n\nfunc Hello() string { return \"hi\" }\n")
	git(t, root, "init", "-b", "main")
	git(t, root, "config", "user.name", "expt")
	git(t, root, "config", "user.email", "expt@example.com")
	git(t, root, "add", ".")
	git(t, root, "commit", "-m", "init")
	git(t, root, "tag", "go-pkgs/v0.0.1")

	consume := filepath.Join(scene, "consume")
	mustMkdir(t, consume)
	mustWrite(t, filepath.Join(consume, "go.mod"), ""+
		"module consume\n\n"+
		"go 1.22\n\n"+
		"require github.com/xhd2015/wrk-seeded-nested/go-pkgs v0.0.1\n")
	mustWrite(t, filepath.Join(consume, "main.go"), ""+
		"package main\n\n"+
		"import (\n"+
		"\t\"fmt\"\n"+
		"\tgopkgs \"github.com/xhd2015/wrk-seeded-nested/go-pkgs\"\n"+
		")\n\n"+
		"func main() { fmt.Print(gopkgs.Hello()) }\n")

	baseEnv := append(os.Environ(),
		"GOMODCACHE="+modCache,
		"GOCACHE="+filepath.Join(scene, "gocache"),
		"GOPATH="+filepath.Join(scene, "gopath"),
		"GOTOOLCHAIN=local",
		"https_proxy=http://127.0.0.1:1",
		"HTTPS_PROXY=http://127.0.0.1:1",
		"http_proxy=http://127.0.0.1:1",
		"HTTP_PROXY=http://127.0.0.1:1",
	)

	if err := tidy.Seeded(context.Background(), tidy.SeededRequest{
		Dir: consume,
		Locals: []seed.Mapping{{
			RepoDir:    mod, // nested module dir
			ModulePath: "github.com/xhd2015/wrk-seeded-nested/go-pkgs",
		}},
		Environ: baseEnv,
	}); err != nil {
		t.Fatalf("Seeded: %v", err)
	}

	extracted := filepath.Join(modCache, "github.com/xhd2015/wrk-seeded-nested/go-pkgs@v0.0.1", "hello.go")
	if _, err := os.Stat(extracted); err != nil {
		t.Fatalf("module not in GOMODCACHE: %v", err)
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(scene, "out"), ".")
	cmd.Dir = consume
	cmd.Env = append(append([]string{}, baseEnv...), "GOPROXY=off", "GOSUMDB=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func chmodTreeWritable(root string) {
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		_ = os.Chmod(path, 0o755)
		return nil
	})
}
