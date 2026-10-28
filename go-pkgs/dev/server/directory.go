package server

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// directoryControl serializes starts for a canonical project directory and
// communicates only with a supervisor that created its private Unix socket.
// In particular, it never infers ownership from a port or a bare PID.
type directoryControl struct {
	lock       *os.File
	socket     string
	listener   net.Listener
	done       chan struct{}
	acceptDone chan struct{}
	handlers   sync.WaitGroup
}

func openDirectoryControl(root string) (*directoryControl, error) {
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve dev directory: %w", err)
	}
	sum := sha256.Sum256([]byte(canonical))
	// Unix-domain socket paths are short on macOS; TMPDIR can be too long.
	base := filepath.Join("/tmp", fmt.Sprintf("dot-pkgs-dev-%d", os.Getuid()))
	if err := os.MkdirAll(base, 0700); err != nil {
		return nil, err
	}
	info, err := os.Lstat(base)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
		return nil, fmt.Errorf("unsafe dev control directory %s", base)
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); !ok || stat.Uid != uint32(os.Getuid()) {
		return nil, fmt.Errorf("dev control directory is not owned by current user: %s", base)
	}
	key := hex.EncodeToString(sum[:16])
	lock, err := os.OpenFile(filepath.Join(base, key+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		lock.Close()
		return nil, err
	}
	return &directoryControl{lock: lock, socket: filepath.Join(base, key+".sock"), done: make(chan struct{})}, nil
}

func (c *directoryControl) unlock() {
	if c.lock != nil {
		_ = syscall.Flock(int(c.lock.Fd()), syscall.LOCK_UN)
		_ = c.lock.Close()
		c.lock = nil
	}
}

func (c *directoryControl) stopExisting(cfg Config) (bool, error) {
	conn, err := net.DialTimeout("unix", c.socket, time.Second)
	if err != nil {
		if _, statErr := os.Lstat(c.socket); os.IsNotExist(statErr) {
			return false, nil
		}
		// A stale socket is safe to remove only when no process accepts it.
		if op, ok := err.(*net.OpError); ok && (op.Err == syscall.ECONNREFUSED || strings.Contains(op.Err.Error(), "connection refused")) {
			if err := os.Remove(c.socket); err != nil && !os.IsNotExist(err) {
				return false, err
			}
			return false, nil
		}
		return false, fmt.Errorf("cannot contact existing dev supervisor: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	message(cfg.Stdout, "90", "Stopping existing dev supervisor for %s...\n", cfg.Root)
	if _, err := fmt.Fprintln(conn, "stop"); err != nil {
		return false, err
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || line != "stopped\n" {
		return false, fmt.Errorf("existing dev supervisor did not stop cleanly: %v", err)
	}
	return true, nil
}

func (c *directoryControl) start(cancel context.CancelFunc) error {
	listener, err := net.Listen("unix", c.socket)
	if err != nil {
		return fmt.Errorf("create dev control socket: %w", err)
	}
	c.listener = listener
	c.acceptDone = make(chan struct{})
	go func() {
		defer close(c.acceptDone)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			c.handlers.Add(1)
			go func() {
				defer c.handlers.Done()
				defer conn.Close()
				_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				line, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil || line != "stop\n" {
					return
				}
				cancel()
				<-c.done
				_, _ = fmt.Fprintln(conn, "stopped")
			}()
		}
	}()
	return nil
}

func (c *directoryControl) close() {
	if c.listener != nil {
		_ = c.listener.Close()
		<-c.acceptDone
	}
	close(c.done)
	c.handlers.Wait()
}
