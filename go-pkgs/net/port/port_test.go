package port

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestFindAvailablePort_returnsStartWhenFree(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen 127.0.0.1:0: %v", err)
	}
	start := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	got, err := FindAvailablePort(start, 1)
	if err != nil {
		t.Fatalf("FindAvailablePort: %v", err)
	}
	if got != start {
		t.Fatalf("got port %d, want %d", got, start)
	}
}

func TestFindAvailablePort_skipsBusy(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen 127.0.0.1:0: %v", err)
	}
	defer ln.Close()
	busy := ln.Addr().(*net.TCPAddr).Port

	got, err := FindAvailablePort(busy, 5)
	if err != nil {
		t.Fatalf("FindAvailablePort: %v", err)
	}
	if got == busy {
		t.Fatalf("returned busy port %d", got)
	}
	if got < busy || got >= busy+5 {
		t.Fatalf("got %d, want in (%d, %d]", got, busy, busy+5)
	}
}

func TestFindAvailablePort_exhausted(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen 127.0.0.1:0: %v", err)
	}
	defer ln.Close()
	busy := ln.Addr().(*net.TCPAddr).Port

	_, err = FindAvailablePort(busy, 1)
	if err == nil {
		t.Fatal("expected error when only candidate is busy")
	}
	wantSub := fmt.Sprintf("starting from %d", busy)
	if got := err.Error(); !strings.Contains(got, wantSub) {
		t.Fatalf("error %q should mention %q", got, wantSub)
	}
}

func TestFindAvailablePort_badMaxAttempts(t *testing.T) {
	_, err := FindAvailablePort(8080, 0)
	if err == nil {
		t.Fatal("expected error for maxAttempts <= 0")
	}
}
