package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func testConfig(root string, backend, frontend int) Config {
	return Config{Root: root, Name: "test", BuildPackage: "./script/dev", BackendPort: backend, FrontendPort: frontend, Stdout: io.Discard, Stderr: io.Discard, BackendHandler: func(context.Context, Backend) (http.Handler, error) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) }), nil
	}}
}

func waitHTTP(t *testing.T, port int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
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

func TestMainReplacesAndKillsSameDirectory(t *testing.T) {
	root := t.TempDir()
	p := freePort(t)
	cfg := testConfig(root, p, 0)
	run := func() <-chan error {
		ch := make(chan error, 1)
		go func() { ch <- Main([]string{"--port", strconv.Itoa(p), "--no-use-air", "--no-open"}, cfg) }()
		return ch
	}
	first := run()
	waitHTTP(t, p)
	second := run()
	select {
	case err := <-first:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("old supervisor did not exit")
	}
	waitHTTP(t, p)
	if err := Main([]string{"--kill"}, cfg); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-second:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("kill did not stop supervisor")
	}
	if err := Main([]string{"--kill"}, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestExplicitEitherPortMakesPairStrict(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	occupied := listener.Addr().(*net.TCPAddr).Port
	cfg := testConfig(t.TempDir(), freePort(t), occupied)
	cfg.Frontend = &Vite{Dir: "unused"}
	if err := Main([]string{"--port", strconv.Itoa(cfg.BackendPort), "--no-open"}, cfg); err == nil {
		t.Fatal("accepted occupied default frontend in explicit mode")
	}
	cfg.BackendPort, cfg.FrontendPort = occupied, freePort(t)
	if err := Main([]string{"--vite-port", strconv.Itoa(cfg.FrontendPort), "--no-open"}, cfg); err == nil {
		t.Fatal("accepted occupied default backend in explicit mode")
	}
	cfg.BackendPort = cfg.FrontendPort
	if err := Main([]string{"--port", strconv.Itoa(cfg.BackendPort), "--vite-port", strconv.Itoa(cfg.FrontendPort)}, cfg); err == nil {
		t.Fatal("accepted equal backend and frontend ports")
	}
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("unrelated listener stopped: %v", err)
	}
	conn.Close()
}
