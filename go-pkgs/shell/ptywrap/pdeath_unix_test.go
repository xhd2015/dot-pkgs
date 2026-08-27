//go:build unix

package ptywrap

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestManagerCloseStopsChild(t *testing.T) {
	mgr := NewManager()
	info, err := mgr.CreateCommand("close-reap", os.TempDir(), []string{"sleep", "60"})
	if err != nil {
		t.Skipf("CreateCommand: %v", err)
	}
	s := mgr.get(info.ID)
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		t.Fatal("missing session process")
	}
	pid := s.cmd.Process.Pid
	mgr.Close()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); err != nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("child pid %d still alive after Manager.Close", pid)
}
