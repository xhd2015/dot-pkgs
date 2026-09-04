package seed_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/seed"
)

func TestVCSRootHTTPS(t *testing.T) {
	u, ok := seed.VCSRootHTTPS("github.com/xhd2015/lib/sub")
	if !ok || u != "https://github.com/xhd2015/lib" {
		t.Fatalf("got %q ok=%v", u, ok)
	}
	if _, ok := seed.VCSRootHTTPS("example.com/lib"); ok {
		t.Fatal("example.com should not be well-known")
	}
}

func TestDownloadFromLocalTag(t *testing.T) {
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
	mustWrite(t, filepath.Join(lib, "go.mod"), "module github.com/xhd2015/wrk-seed-expt\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(lib, "hello.go"), "package seedexpt\n\nfunc Hello() string { return \"hi\" }\n")
	git(t, lib, "init", "-b", "main")
	git(t, lib, "config", "user.name", "expt")
	git(t, lib, "config", "user.email", "expt@example.com")
	git(t, lib, "add", ".")
	git(t, lib, "commit", "-m", "init")
	git(t, lib, "tag", "v0.0.1")
	head := strings.TrimSpace(gitOut(t, lib, "rev-parse", "HEAD"))

	t.Setenv("GOMODCACHE", modCache)
	t.Setenv("GOCACHE", filepath.Join(scene, "gocache"))
	t.Setenv("GOPATH", filepath.Join(scene, "gopath"))
	t.Setenv("GOTOOLCHAIN", "local")
	// Poison HTTPS so a missed insteadOf cannot succeed via GitHub.
	t.Setenv("https_proxy", "http://127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")
	t.Setenv("http_proxy", "http://127.0.0.1:1")
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")

	res, err := seed.Download(context.Background(), seed.Request{
		RepoDir: lib,
		Modules: []seed.Module{{Path: "github.com/xhd2015/wrk-seed-expt", Version: "v0.0.1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Modules) != 1 {
		t.Fatalf("modules: %+v", res.Modules)
	}
	mr := res.Modules[0]
	if mr.Err != nil {
		t.Fatalf("module err: %v", mr.Err)
	}
	if mr.Skipped {
		t.Fatal("unexpected skip")
	}
	if mr.Hash != head {
		t.Fatalf("hash %s want %s", mr.Hash, head)
	}
	if mr.Zip == "" {
		t.Fatal("empty zip path")
	}
	if _, err := os.Stat(mr.Zip); err != nil {
		t.Fatalf("zip missing: %v", err)
	}

	// Later list with GOPROXY=off, no insteadOf.
	consume := filepath.Join(scene, "consume")
	mustMkdir(t, consume)
	mustWrite(t, filepath.Join(consume, "go.mod"), "module consume\n\ngo 1.22\n")
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Version}}", "github.com/xhd2015/wrk-seed-expt@v0.0.1")
	cmd.Dir = consume
	cmd.Env = append(os.Environ(),
		"GOMODCACHE="+modCache,
		"GOCACHE="+filepath.Join(scene, "gocache"),
		"GOPATH="+filepath.Join(scene, "gopath"),
		"GOTOOLCHAIN=local",
		"GOPROXY=off",
		"GOSUMDB=off",
		"https_proxy=http://127.0.0.1:1",
		"HTTPS_PROXY=http://127.0.0.1:1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(string(out)); got != "v0.0.1" {
		t.Fatalf("go list version %q", got)
	}
}

func TestDownloadNestedSubdirTag(t *testing.T) {
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
	mustWrite(t, filepath.Join(lib, "go.mod"), "module github.com/xhd2015/wrk-seed-expt\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(lib, "root.go"), "package seedexpt\n")
	sub := filepath.Join(lib, "sub")
	mustMkdir(t, sub)
	mustWrite(t, filepath.Join(sub, "go.mod"), "module github.com/xhd2015/wrk-seed-expt/sub\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(sub, "sub.go"), "package sub\n\nfunc N() int { return 7 }\n")
	git(t, lib, "init", "-b", "main")
	git(t, lib, "config", "user.name", "expt")
	git(t, lib, "config", "user.email", "expt@example.com")
	git(t, lib, "add", ".")
	git(t, lib, "commit", "-m", "init")
	git(t, lib, "tag", "v0.0.1")
	git(t, lib, "tag", "sub/v0.0.1")
	subHead := strings.TrimSpace(gitOut(t, lib, "rev-parse", "HEAD"))

	t.Setenv("GOMODCACHE", modCache)
	t.Setenv("GOCACHE", filepath.Join(scene, "gocache"))
	t.Setenv("GOPATH", filepath.Join(scene, "gopath"))
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("https_proxy", "http://127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")

	res, err := seed.Download(context.Background(), seed.Request{
		RepoDir: lib,
		Modules: []seed.Module{{Path: "github.com/xhd2015/wrk-seed-expt/sub", Version: "v0.0.1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	mr := res.Modules[0]
	if mr.Err != nil {
		t.Fatal(mr.Err)
	}
	if mr.Hash != subHead {
		t.Fatalf("hash %s want %s", mr.Hash, subHead)
	}
	if mr.Ref != "refs/tags/sub/v0.0.1" {
		t.Fatalf("ref %q", mr.Ref)
	}
}

func TestDownloadSkipsUnknownHost(t *testing.T) {
	res, err := seed.Download(context.Background(), seed.Request{
		RepoDir: t.TempDir(),
		Modules: []seed.Module{{Path: "example.com/lib", Version: "v0.0.1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Modules[0].Skipped || res.Modules[0].Err != nil {
		t.Fatalf("%+v", res.Modules[0])
	}
}

// Nested Go module under a parent git root: RepoDir is the module dir
// (…/go-pkgs), not the git toplevel — OverlayEnv must resolve toplevel.
func TestDownloadNestedModuleDirAsRepoDir(t *testing.T) {
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
	mustWrite(t, filepath.Join(mod, "go.mod"), "module github.com/xhd2015/wrk-seed-nested/go-pkgs\n\ngo 1.22\n")
	mustWrite(t, filepath.Join(mod, "hello.go"), "package gopkgs\n\nfunc Hello() string { return \"hi\" }\n")
	git(t, root, "init", "-b", "main")
	git(t, root, "config", "user.name", "expt")
	git(t, root, "config", "user.email", "expt@example.com")
	git(t, root, "add", ".")
	git(t, root, "commit", "-m", "init")
	git(t, root, "tag", "go-pkgs/v0.0.1")
	head := strings.TrimSpace(gitOut(t, root, "rev-parse", "HEAD"))

	t.Setenv("GOMODCACHE", modCache)
	t.Setenv("GOCACHE", filepath.Join(scene, "gocache"))
	t.Setenv("GOPATH", filepath.Join(scene, "gopath"))
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("https_proxy", "http://127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:1")

	res, err := seed.Download(context.Background(), seed.Request{
		RepoDir: mod, // nested module dir, not git root
		Modules: []seed.Module{{Path: "github.com/xhd2015/wrk-seed-nested/go-pkgs", Version: "v0.0.1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	mr := res.Modules[0]
	if mr.Err != nil {
		t.Fatal(mr.Err)
	}
	if mr.Skipped {
		t.Fatal("unexpected skip")
	}
	if mr.Hash != head {
		t.Fatalf("hash %s want %s", mr.Hash, head)
	}
	if mr.Ref != "refs/tags/go-pkgs/v0.0.1" {
		t.Fatalf("ref %q", mr.Ref)
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
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}

// chmodTreeWritable makes a Go module cache tree removable (entries are 0444).
func chmodTreeWritable(root string) {
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		mode := info.Mode()
		if mode.IsDir() {
			_ = os.Chmod(path, mode|0o700)
		} else {
			_ = os.Chmod(path, mode|0o600)
		}
		return nil
	})
}
