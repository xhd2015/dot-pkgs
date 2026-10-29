package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuffer is a concurrency-safe buffer for supervisor output.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// runProject writes a minimal module whose script/run child serves through the
// shared one-shot contract, replacing go-pkgs with this package's checkout.
func runProject(t *testing.T) string {
	t.Helper()
	module := findModuleRoot(t)
	root := t.TempDir()
	files := map[string]string{
		"go.mod": fmt.Sprintf("module runtest\n\ngo 1.25\n\nrequire github.com/xhd2015/dot-pkgs/go-pkgs v0.0.0\n\nreplace github.com/xhd2015/dot-pkgs/go-pkgs => %s\n", filepath.ToSlash(module)),
		"go.sum": readModuleSum(t),
		"script/run/main.go": `package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	devserver "github.com/xhd2015/dot-pkgs/go-pkgs/dev/server"
)

func main() {
	err := devserver.RunMain(os.Args[1:], devserver.Config{
		Name: "runtest", Root: ".", BuildPackage: "./script/run",
		BackendHandler: func(context.Context, devserver.Backend) (http.Handler, error) {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "child-ok") }), nil
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`,
	}
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// The pruned module graph requires the transitive requires recorded in the
	// temp go.mod; tidy resolves them from the module cache only.
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = root
	tidy.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOPROXY=off")
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}
	return root
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	// Tests run in the package directory; walk up to the module root.
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the test directory")
		}
		dir = parent
	}
}

func readModuleSum(t *testing.T) string {
	t.Helper()
	sum, err := os.ReadFile(filepath.Join(findModuleRoot(t), "go.sum"))
	if err != nil {
		t.Fatal(err)
	}
	return string(sum)
}

func waitReadyHTTP(t *testing.T, port int, within time.Duration) {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/")
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatal("server did not become ready")
}

func TestRunMainBuildsServesAndStopsCleanly(t *testing.T) {
	root := runProject(t)
	ui := filepath.Join(root, "ui")
	if err := os.MkdirAll(filepath.Join(ui, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := freePort(t)
	cfg := Config{Root: root, Name: "runtest", BuildPackage: "./script/run", BackendPort: p,
		StartupTimeout: 60 * time.Second, Stdout: io.Discard, Stderr: io.Discard,
		FrontendBuild: &FrontendBuild{Dir: "ui", Command: []string{"sh", "-c", "mkdir -p dist && echo built > dist/index.html"}},
		BackendHandler: func(context.Context, Backend) (http.Handler, error) {
			return http.NotFoundHandler(), nil
		},
	}
	done := make(chan error, 1)
	go func() { done <- RunMain([]string{"--port", strconv.Itoa(p), "--no-open"}, cfg) }()
	waitReadyHTTP(t, p, 60*time.Second)
	if _, err := os.Stat(filepath.Join(ui, "dist", "index.html")); err != nil {
		t.Fatalf("frontend not built: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "tmp")); !os.IsNotExist(err) {
		t.Fatalf("repo artifacts left: %v", err)
	}
	if err := RunMain([]string{"--kill"}, cfg); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown deadlocked")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(p))
	if err != nil {
		t.Fatalf("backend port not released: %v", err)
	}
	ln.Close()
}

func TestRunAndDevKillEachOther(t *testing.T) {
	// A run supervisor replaces a running dev supervisor.
	root := runProject(t)
	p := freePort(t)
	devDone := make(chan error, 1)
	go func() {
		devDone <- Main([]string{"--port", strconv.Itoa(p), "--no-use-air", "--no-open"}, testConfig(root, p, 0))
	}()
	waitReadyHTTP(t, p, 5*time.Second)
	runCfg := testConfig(root, p, 0)
	runCfg.StartupTimeout = 60 * time.Second
	runCfg.FrontendBuild = nil
	if err := RunMain([]string{"--kill"}, runCfg); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-devDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("kill did not stop dev supervisor")
	}
	// A dev supervisor replaces a running run supervisor.
	p2 := freePort(t)
	runCfg = testConfig(root, p2, 0)
	runCfg.BuildPackage = "./script/run"
	runCfg.StartupTimeout = 60 * time.Second
	var runLog syncBuffer
	runCfg.Stdout, runCfg.Stderr = &runLog, &runLog
	defer func() {
		if t.Failed() {
			t.Logf("run supervisor output:\n%s", runLog.String())
		}
	}()
	runDone := make(chan error, 1)
	go func() { runDone <- RunMain([]string{"--port", strconv.Itoa(p2), "--no-open"}, runCfg) }()
	waitReadyHTTP(t, p2, 60*time.Second)
	if err := Main([]string{"--kill"}, testConfig(root, p2, 0)); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("kill did not stop run supervisor")
	}
}

func TestRunMainKillWithoutSupervisor(t *testing.T) {
	if err := RunMain([]string{"--kill"}, testConfig(t.TempDir(), freePort(t), 0)); err != nil {
		t.Fatal(err)
	}
}

func TestRunMainRejectsDevFlags(t *testing.T) {
	for _, args := range [][]string{{"--vite-port", "5199"}, {"--vite-port=5199"}, {"--use-air"}, {"--no-use-air"}, {"--no-air"}} {
		err := RunMain(args, testConfig(t.TempDir(), freePort(t), 0))
		if err == nil {
			t.Fatalf("%v accepted", args)
		}
		if !strings.Contains(err.Error(), "dev flag") {
			t.Fatalf("%v: %v", args, err)
		}
	}
}

func TestRunMainFrontendBuildFailureStopsBeforeBackend(t *testing.T) {
	root := runProject(t)
	ui := filepath.Join(root, "ui")
	if err := os.MkdirAll(filepath.Join(ui, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Root: root, Name: "runtest", BuildPackage: "./script/run", BackendPort: freePort(t),
		StartupTimeout: 60 * time.Second, Stdout: io.Discard, Stderr: io.Discard,
		FrontendBuild: &FrontendBuild{Dir: "ui", Command: []string{"false"}},
		BackendHandler: func(context.Context, Backend) (http.Handler, error) {
			return http.NotFoundHandler(), nil
		},
	}
	err := RunMain([]string{"--port", strconv.Itoa(cfg.BackendPort), "--no-open"}, cfg)
	if err == nil || !strings.Contains(err.Error(), "frontend build") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "tmp")); !os.IsNotExist(err) {
		t.Fatalf("repo artifacts left: %v", err)
	}
}
