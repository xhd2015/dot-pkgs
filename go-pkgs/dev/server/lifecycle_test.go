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

func TestPortSelectionIndependentPerFlag(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	occupied := listener.Addr().(*net.TCPAddr).Port
	if occupied == 65535 {
		t.Skip("no higher port to test automatic selection")
	}
	free := freePort(t)
	cfg := testConfig(t.TempDir(), free, occupied)
	cfg.Frontend = &Vite{Dir: "unused"}

	// Only the backend is explicit: Vite must skip the occupied default.
	chosen, err := selectPorts(cfg, options{port: free})
	if err != nil || chosen.port != free || chosen.vitePort == occupied {
		t.Fatalf("explicit backend: %+v, %v", chosen, err)
	}
	// Only Vite is explicit: the backend must skip its occupied default.
	cfg.BackendPort, cfg.FrontendPort = occupied, free
	chosen, err = selectPorts(cfg, options{vitePort: free})
	if err != nil || chosen.vitePort != free || chosen.port == occupied || chosen.port == free {
		t.Fatalf("explicit Vite: %+v, %v", chosen, err)
	}
	// Both automatic, and zero-valued flags, follow the same search path.
	chosen, err = selectPorts(cfg, options{port: 0, vitePort: 0})
	if err != nil || chosen.port == occupied || chosen.port == chosen.vitePort {
		t.Fatalf("automatic ports: %+v, %v", chosen, err)
	}
	// A Vite port explicitly equal to the backend default is reserved for Vite.
	if free < 65535 {
		cfg.BackendPort = free
		chosen, err = selectPorts(cfg, options{vitePort: free})
		if err != nil || chosen.port == free || chosen.vitePort != free {
			t.Fatalf("reserved Vite port: %+v, %v", chosen, err)
		}
	}
	// Explicitly requested occupied ports and duplicate ports must fail.
	for _, opts := range []options{{port: occupied}, {vitePort: occupied}, {port: free, vitePort: free}} {
		if _, err := selectPorts(cfg, opts); err == nil {
			t.Fatalf("accepted invalid explicit ports: %+v", opts)
		}
	}
	otherFree := freePort(t)
	for otherFree == free {
		otherFree = freePort(t)
	}
	chosen, err = selectPorts(cfg, options{port: free, vitePort: otherFree})
	if err != nil || chosen.port == chosen.vitePort {
		t.Fatalf("two distinct explicit ports: %+v, %v", chosen, err)
	}
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("unrelated listener stopped: %v", err)
	}
	conn.Close()
}
