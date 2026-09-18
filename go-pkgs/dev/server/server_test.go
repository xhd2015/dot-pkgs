package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func fmtPort(p int) string { return strconv.Itoa(p) }

func TestOwnedShutdownRunsCleanupAndRejectsWrongIdentity(t *testing.T) {
	p := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var cleaned atomic.Int32
	cfg := Config{BackendHandler: func(context.Context, Backend) (http.Handler, error) { return http.NotFoundHandler(), nil }, OnListen: func(Backend) (func(), error) { return func() { cleaned.Add(1) }, nil }}
	done := make(chan error, 1)
	go func() { done <- serve(ctx, cfg, options{port: p}, "this-run") }()
	url := "http://127.0.0.1:" + fmtPort(p)
	readyCtx, stop := context.WithTimeout(ctx, 3*time.Second)
	defer stop()
	if err := waitReady(readyCtx, url, "this-run"); err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("POST", url+"/__dev/stop", nil)
	req.Header.Set("X-Dev-Run-ID", "another-run")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden || cleaned.Load() != 0 {
		t.Fatal("foreign invocation stopped backend")
	}
	stopOwnedBackend(url, "this-run")
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("backend not stopped")
	}
	if cleaned.Load() != 1 {
		t.Fatalf("cleanup count %d", cleaned.Load())
	}
}

func TestReadinessRejectsForeignAndStaleServers(t *testing.T) {
	for _, response := range []string{"pong", "old-run"} {
		t.Run(response, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, response) }))
			defer srv.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			if err := waitReady(ctx, srv.URL, "new-run"); err == nil {
				t.Fatal("foreign readiness accepted")
			}
		})
	}
}

func TestReadinessAcceptsIdentityAndRejectsRedirect(t *testing.T) {
	srv := httptest.NewServer(identityHandler("new-run", http.NotFoundHandler()))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := waitReady(ctx, srv.URL, "new-run"); err != nil {
		t.Fatal(err)
	}
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, srv.URL+readyPath, 302) }))
	defer redirect.Close()
	short, stop := context.WithTimeout(ctx, 200*time.Millisecond)
	defer stop()
	if err := waitReady(short, redirect.URL, "new-run"); err == nil {
		t.Fatal("followed readiness redirect")
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func TestPortSelectionPreservesExistingListener(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	p := ln.Addr().(*net.TCPAddr).Port
	cfg := Config{Stderr: io.Discard}
	if _, err := selectPort(p, p, 0, cfg, "backend"); err == nil {
		t.Fatal("accepted occupied explicit port")
	}
	chosen, err := selectPort(0, p, 0, cfg, "backend")
	if err != nil {
		t.Fatal(err)
	}
	if chosen == p {
		t.Fatal("selected occupied default")
	}
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	chosen, err = selectPort(0, chosen, chosen, cfg, "frontend")
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoAirLifecycleAndDiscoveryCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var cleaned atomic.Bool
	p := freePort(t)
	cfg := Config{Name: "test", Root: t.TempDir(), BackendPort: p, StartupTimeout: time.Second, Stdout: io.Discard, Stderr: io.Discard,
		BackendHandler: func(context.Context, Backend) (http.Handler, error) { return http.NotFoundHandler(), nil },
		OnListen:       func(Backend) (func(), error) { return func() { cleaned.Store(true) }, nil },
	}
	done := make(chan error, 1)
	go func() { done <- supervise(ctx, cfg, options{noAir: true, noOpen: true}) }()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://127.0.0.1:" + fmtPort(p) + readyPath)
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown deadlocked")
	}
	if !cleaned.Load() {
		t.Fatal("discovery cleanup not called")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:"+fmtPort(p))
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()
}

func TestBindRaceDoesNotInitializeApplication(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	err = serve(context.Background(), Config{BackendHandler: func(context.Context, Backend) (http.Handler, error) {
		t.Fatal("initialized despite bind failure")
		return nil, nil
	}}, options{port: ln.Addr().(*net.TCPAddr).Port}, "run")
	if err == nil || !strings.Contains(err.Error(), "bind") {
		t.Fatalf("got %v", err)
	}
}

func TestCancellationDuringInstallStopsProcess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := runCommand(ctx, t.TempDir(), []string{"sh", "-c", "sleep 30"}, Config{Stdout: io.Discard, Stderr: io.Discard})
	if err == nil {
		t.Fatal("expected cancellation")
	}
}

func TestFrontendExitCleansArtifacts(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ui", "node_modules"), 0755); err != nil {
		t.Fatal(err)
	}
	cfg := Config{Root: root, BackendPort: freePort(t), FrontendPort: freePort(t), StartupTimeout: time.Second, Stdout: io.Discard, Stderr: io.Discard,
		Frontend: &Vite{Dir: "ui", Command: []string{"false"}},
	}
	if err := supervise(context.Background(), cfg, options{noOpen: true}); err == nil {
		t.Fatal("expected frontend failure")
	}
	entries, err := os.ReadDir(filepath.Join(root, "tmp"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("artifacts left: %v %v", entries, err)
	}
}
