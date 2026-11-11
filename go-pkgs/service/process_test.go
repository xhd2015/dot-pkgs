package service

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestParseProcStatState(t *testing.T) {
	cases := []struct {
		stat  string
		state byte
		ok    bool
	}{
		{"123 (bash) S 1 123 123 0 -1", 'S', true},
		{"9 (a b) Z 1 9 9 0 -1", 'Z', true},
		{"42 (weird) name) R 1 42", 'R', true},
		{"broken", 0, false},
		{"1 () ", 0, false},
		{"", 0, false},
	}
	for _, tc := range cases {
		state, ok := ParseProcStatState([]byte(tc.stat))
		if state != tc.state || ok != tc.ok {
			t.Fatalf("ParseProcStatState(%q) = %q, %v, want %q, %v", tc.stat, state, ok, tc.state, tc.ok)
		}
	}
}

func TestCommandMatches(t *testing.T) {
	cases := []struct {
		recorded string
		live     string
		want     bool
	}{
		{"sleep 300", "bash -lc sleep 300", true},
		{"sleep 300", "sleep 300", true},
		{"", "sleep 300", false},
		{"sleep 300", "", false},
		{"other", "sleep 300", false},
	}
	for _, tc := range cases {
		if got := CommandMatches(tc.recorded, tc.live); got != tc.want {
			t.Fatalf("CommandMatches(%q, %q) = %v, want %v", tc.recorded, tc.live, got, tc.want)
		}
	}
}

func TestAliveOwnProcess(t *testing.T) {
	if !Alive(os.Getpid()) {
		t.Fatalf("Alive(self) = false")
	}
	for _, pid := range []int{0, -1} {
		if Alive(pid) {
			t.Fatalf("Alive(%d) = true", pid)
		}
	}
	dead := exec.Command("bash", "-c", "exit 0")
	if err := dead.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	pid := dead.Process.Pid
	_ = dead.Wait()
	if Alive(pid) {
		t.Fatalf("Alive(%d) = true after the child was reaped", pid)
	}
}

func TestProcessCommandLineOfCurrentProcess(t *testing.T) {
	line := ProcessCommandLine(os.Getpid())
	if line == "" {
		t.Fatalf("ProcessCommandLine(self) is empty")
	}
	if !CommandMatches("go", line) && !CommandMatches("service.test", line) {
		t.Logf("ProcessCommandLine(self) = %q", line)
	}
	if got := ProcessCommandLine(0); got != "" {
		t.Fatalf("ProcessCommandLine(0) = %q, want empty", got)
	}
}

func TestGroupIDRejectsInvalidPID(t *testing.T) {
	if _, err := GroupID(0); err == nil {
		t.Fatalf("GroupID(0) error = nil")
	}
}

func TestStopProcessKillsWholeGroup(t *testing.T) {
	cmd := exec.Command("bash", "-c", "sleep 300 & wait")
	cmd.SysProcAttr = procAttr(false)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	pid := cmd.Process.Pid
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	child := firstChildPID(t, pid)
	if child == 0 {
		t.Skip("pgrep not available; cannot observe the child")
	}
	if err := StopProcessWithTimeout(pid, 3*time.Second, 2*time.Second); err != nil {
		t.Fatalf("StopProcessWithTimeout() error = %v", err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("group leader did not exit")
	}
	if Alive(child) {
		t.Fatalf("child %d survived the group stop", child)
	}
}

func TestKillProcessGroupNow(t *testing.T) {
	cmd := exec.Command("sleep", "300")
	cmd.SysProcAttr = procAttr(false)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	pid := cmd.Process.Pid
	go func() { _ = cmd.Wait() }()
	KillProcessGroupNow(pid)
	if Alive(pid) {
		t.Fatalf("pid %d survived KillProcessGroupNow", pid)
	}
}

func TestWaitingPIDsFindsListenerAndKillPortOwner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no lsof")
	}
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not installed")
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	port := freePort(t)
	script := fmt.Sprintf("import socket,time;s=socket.socket();s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1);s.bind(('127.0.0.1',%d));s.listen(1);time.sleep(300)", port)
	cmd := exec.Command("python3", "-c", script)
	cmd.SysProcAttr = procAttr(false)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	go func() { _ = cmd.Wait() }()
	defer func() { _ = cmd.Process.Kill() }()

	deadline := time.Now().Add(5 * time.Second)
	for {
		pids, err := ListeningPIDs(port)
		if err != nil {
			t.Fatalf("ListeningPIDs() error = %v", err)
		}
		if len(pids) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("port %d never showed a listener", port)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := KillPortOwner(port); err != nil {
		t.Fatalf("KillPortOwner() error = %v", err)
	}
	pids, err := ListeningPIDs(port)
	if err != nil {
		t.Fatalf("ListeningPIDs() error = %v", err)
	}
	if len(pids) != 0 {
		t.Fatalf("listener pids after kill = %v, want none", pids)
	}
}

func TestListeningPIDsUnusedPort(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not installed")
	}
	pids, err := ListeningPIDs(freePort(t))
	if err != nil {
		t.Fatalf("ListeningPIDs() error = %v", err)
	}
	if len(pids) != 0 {
		t.Fatalf("ListeningPIDs() = %v, want none", pids)
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func firstChildPID(t *testing.T, pid int) int {
	t.Helper()
	if _, err := exec.LookPath("pgrep"); err != nil {
		return 0
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		out, err := exec.Command("pgrep", "-P", strconv.Itoa(pid)).Output()
		if err == nil {
			fields := strings.Fields(string(out))
			if len(fields) > 0 {
				child, convErr := strconv.Atoi(fields[0])
				if convErr == nil {
					return child
				}
			}
		}
		if time.Now().After(deadline) {
			return 0
		}
		time.Sleep(50 * time.Millisecond)
	}
}
