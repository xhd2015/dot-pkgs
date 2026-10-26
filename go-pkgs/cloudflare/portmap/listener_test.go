package portmap

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrivateUnixRelayIsUniqueAndRemovedOnClose(t *testing.T) {
	p := &fakePublisher{}
	m, err := New(Options{Publisher: p, UseUnixSocket: true})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	path := strings.TrimPrefix(m.origin, "unix:")
	if path == m.origin {
		t.Fatalf("not a unix origin: %s", m.origin)
	}
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0700 {
		t.Fatal("relay directory must be private")
	}
	client := &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", path)
	}}}
	defer client.CloseIdleConnections()
	resp, err := client.Get("http://inactive.example.com/")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 410 {
		t.Fatal(resp.StatusCode)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("socket still exists: %v", err)
	}
	next, err := New(Options{Publisher: p, UseUnixSocket: true})
	if err != nil {
		t.Fatal(err)
	}
	defer next.Close()
	if next.origin == m.origin {
		t.Fatal("stale connector origin was reused")
	}
}
