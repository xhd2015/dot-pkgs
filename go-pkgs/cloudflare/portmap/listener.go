package portmap

import (
	"net"
	"os"
	"path/filepath"
)

// A private, unique Unix socket is crash-safe: a surviving connector can never
// accidentally publish an unrelated app that reuses an ephemeral TCP port.
// Stale sockets have no listener and are never reused by later managers.
func listenRelay(useUnix bool) (net.Listener, string, func(), error) {
	if !useUnix {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return nil, "", nil, err
		}
		return listener, "http://" + listener.Addr().String(), func() {}, nil
	}
	parent := os.TempDir()
	// macOS uses a long per-user TMPDIR but sockaddr_un allows only 104 bytes.
	if len(filepath.Join(parent, "portmap-relay-1234567890", "http.sock")) >= 100 {
		parent = "/tmp"
	}
	dir, err := os.MkdirTemp(parent, "portmap-relay-")
	if err != nil {
		return nil, "", nil, err
	}
	path := filepath.Join(dir, "http.sock")
	cleanup := func() { _ = os.Remove(dir) } // only our newly created, empty directory
	listener, err := net.Listen("unix", path)
	if err != nil {
		cleanup()
		return nil, "", nil, err
	}
	return listener, "unix:" + path, cleanup, nil
}
