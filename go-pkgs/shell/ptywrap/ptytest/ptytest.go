package ptytest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
)

// Request is the doctest harness request for ptywrap server tests.
type Request struct {
	Phase string

	Command    []string
	Cwd        string
	Name       string
	SessionID  string
	AttachMode string

	WSInput      string
	ResizeCols   int
	ResizeRows   int
	ExpectMarker string

	RenameTo string

	// WSCloseCode is the WebSocket close code used by lifecycle leak phases
	// (e.g. 1000 normal close, 4000 delete-on-close).
	WSCloseCode int

	// RepeatCount is used by multi-create leak phases (default 5).
	RepeatCount int

	ServerBase string
}

// Response is the doctest harness response for ptywrap server tests.
type Response struct {
	SessionID       string
	Sessions        []ptywrap.SessionInfo
	WSOutput        string
	ReconnectOutput string
	IsTTY           bool
	PTYCols         int
	PTYRows         int
	HTTPStatus      int
	CreateBody      map[string]interface{}

	// ProcessAlive is true if the session child PID still exists after the phase.
	ProcessAlive bool
	// SessionListed is true if SessionID is still present in GET /sessions.
	SessionListed bool
	// RunningProcessCount is how many tracked child PIDs are still alive.
	RunningProcessCount int
	// SessionCount is Total (or len) from GET /sessions after the phase.
	SessionCount int

	// SnapshotCount is how many attach_mode=snapshot reads returned usable bytes.
	SnapshotCount int

	// CloseCode is the WebSocket close code observed when the server ends the
	// attach socket after the shell exits (expected 1000 after clean-close fix).
	// Zero means no close was observed (timeout / harness error path).
	CloseCode int
	// AttachErr is the error string from client Attach/Wait after shell exit.
	// Empty string means nil error (success).
	AttachErr string
	// SessionStatus is the REST status of SessionID after the phase (e.g. "exited").
	SessionStatus string
}

// Run executes a ptywrap server doctest phase.
func Run(t *testing.T, req *Request) (*Response, error) {
	if req.ServerBase == "" {
		return nil, fmt.Errorf("ServerBase not set; root Setup must start server")
	}
	switch req.Phase {
	case "spawn-shell":
		return runSpawnShellTTY(t, req)
	case "spawn-cmd":
		return runSpawnCommandOutput(t, req)
	case "ws-input":
		return runWSInputRoundTrip(t, req)
	case "ws-resize":
		return runWSResize(t, req)
	case "ws-reconnect":
		return runWSReconnectScrollback(t, req)
	case "ws-create-on-connect":
		return runWSCreateOnConnect(t, req)
	case "rest-create":
		return runRESTCreate(t, req)
	case "rest-rename":
		return runRESTRename(t, req)
	case "lifecycle-exited":
		return runExitedStaysListed(t, req)
	case "lifecycle-writer-close":
		return runWriterCloseLifecycle(t, req)
	case "lifecycle-multi-create-orphan":
		return runMultiCreateOnConnectOrphan(t, req)
	case "snapshot-multi-keeps-child":
		return runSnapshotMultiKeepsChild(t, req)
	case "shell-exit-ws-close-code":
		return runShellExitWSCloseCode(t, req)
	case "shell-exit-attach-wait":
		return runShellExitAttachWait(t, req)
	case "shell-exit-marker-without-close":
		return runShellExitMarkerWithoutClose(t, req)
	case "shell-exit-hard-drop-without-marker":
		return runShellExitHardDropWithoutMarker(t, req)
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}

// StartTestServer starts an httptest server with ptywrap handlers.
func StartTestServer(t *testing.T) (base string, cleanup func()) {
	t.Helper()
	mux := http.NewServeMux()
	ptywrap.RegisterAPI(mux)
	srv := httptest.NewServer(mux)
	return srv.URL, srv.Close
}

// AbsTempDir returns an absolute temp directory for tests.
func AbsTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func runSpawnShellTTY(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	id, err := createSessionREST(t, req.ServerBase, nil, req.Cwd, req.Name)
	if err != nil {
		return nil, err
	}
	out, err := wsRunCommand(t, req.ServerBase, id, "python3 -c \"import sys,os; print('tty=' + ('1' if os.isatty(1) else '0'))\"\n")
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.WSOutput = out
	resp.IsTTY = strings.Contains(out, "tty=1")
	return resp, nil
}

func runSpawnCommandOutput(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	cmd := req.Command
	if len(cmd) == 0 {
		cmd = []string{"echo", "hello"}
	}
	id, err := createSessionREST(t, req.ServerBase, cmd, req.Cwd, req.Name)
	if err != nil {
		return nil, err
	}
	out, err := wsCollectOutput(t, req.ServerBase, id, 2*time.Second)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.WSOutput = out
	return resp, nil
}

func runWSInputRoundTrip(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	id, err := createSessionREST(t, req.ServerBase, []string{"cat"}, "", "")
	if err != nil {
		return nil, err
	}
	input := req.WSInput
	if input == "" {
		input = "roundtrip-marker\n"
	}
	out, err := wsRunCommand(t, req.ServerBase, id, input)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.WSOutput = out
	return resp, nil
}

func runWSResize(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	// Interactive default shell so `stty size` is executed (a bare `sleep`
	// ignores keystrokes on the PTY slave).
	id, err := createSessionREST(t, req.ServerBase, nil, "", "")
	if err != nil {
		return nil, err
	}
	cols, rows := req.ResizeCols, req.ResizeRows
	if cols <= 0 {
		cols = 100
	}
	if rows <= 0 {
		rows = 40
	}
	out, err := wsResizeAndCollect(t, req.ServerBase, id, cols, rows, 3*time.Second)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.WSOutput = out
	resp.PTYCols, resp.PTYRows = parseSttySize(out)
	return resp, nil
}

func runWSReconnectScrollback(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	id, err := createSessionREST(t, req.ServerBase, []string{"sh", "-c", "printf scrollback-marker; sleep 30"}, "", "")
	if err != nil {
		return nil, err
	}
	firstOut, err := wsCollectOutput(t, req.ServerBase, id, 2*time.Second)
	if err != nil {
		return nil, err
	}
	reOut, err := wsAttachOnly(t, req.ServerBase, id, req.AttachMode, 2*time.Second)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.WSOutput = firstOut
	resp.ReconnectOutput = reOut
	return resp, nil
}

func runWSCreateOnConnect(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	cwd := req.Cwd
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	name := req.Name
	if name == "" {
		name = "compat-shell"
	}
	id, out, err := wsCreateOnConnect(t, req.ServerBase, name, cwd)
	if err != nil {
		return nil, err
	}
	sessions, err := listSessions(t, req.ServerBase)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.WSOutput = out
	resp.Sessions = sessions
	return resp, nil
}

func runRESTCreate(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	body, status, err := postCreateSession(t, req.ServerBase, req.Command, req.Cwd, req.Name)
	if err != nil {
		return nil, err
	}
	resp.HTTPStatus = status
	resp.CreateBody = body
	if id, ok := body["id"].(string); ok {
		resp.SessionID = id
	}
	return resp, nil
}

func runRESTRename(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	id, err := createSessionREST(t, req.ServerBase, []string{"sleep", "60"}, "", "before-rename")
	if err != nil {
		return nil, err
	}
	newName := req.RenameTo
	if newName == "" {
		newName = "after-rename"
	}
	if err := patchRenameSession(t, req.ServerBase, id, newName); err != nil {
		return nil, err
	}
	sessions, err := listSessions(t, req.ServerBase)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.Sessions = sessions
	return resp, nil
}

func runExitedStaysListed(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	cmd := req.Command
	if len(cmd) == 0 {
		cmd = []string{"true"}
	}
	id, err := createSessionREST(t, req.ServerBase, cmd, "", "exit-test")
	if err != nil {
		return nil, err
	}
	if _, err := wsWaitDone(t, req.ServerBase, id, 10*time.Second); err != nil {
		return nil, err
	}
	sessions, err := listSessions(t, req.ServerBase)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.Sessions = sessions
	return resp, nil
}

// runSnapshotMultiKeepsChild starts a long-lived child, performs N short
// attach_mode=snapshot reads (closing each socket), and reports whether the
// child is still alive. Snapshot role must not claim writer / stopChild on
// disconnect — writer attach_mode=screen would kill the child and break
// multi-poll FetchStatus-style loops.
func runSnapshotMultiKeepsChild(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	n := req.RepeatCount
	if n <= 0 {
		n = 3
	}
	attachMode := strings.TrimSpace(req.AttachMode)
	if attachMode == "" {
		attachMode = "snapshot"
	}

	marker := fmt.Sprintf("ptywrap-snap-%d-%d", os.Getpid(), time.Now().UnixNano())
	// Print a marker then sleep so scrollback is non-empty for snapshot frames.
	cmd := []string{"sh", "-c", "printf 'SNAP-MARKER\\n'; sleep 3600"}
	beforeSleep := snapshotSleep3600PIDs()
	id, err := createSessionREST(t, req.ServerBase, cmd, "", marker)
	if err != nil {
		return nil, err
	}
	pid, err := findNewSleepPID(t, beforeSleep)
	if err != nil {
		return nil, fmt.Errorf("session %s: %w", id, err)
	}

	// Wait briefly for the printf to land in scrollback before snapshotting.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		out, snapErr := wsAttachOnly(t, req.ServerBase, id, attachMode, 500*time.Millisecond)
		if snapErr == nil && strings.Contains(out, "SNAP-MARKER") {
			resp.WSOutput = out
			resp.SnapshotCount = 1
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if resp.SnapshotCount == 0 {
		// Still run the remaining attaches so ProcessAlive reflects full multi-snapshot stress.
		out, snapErr := wsAttachOnly(t, req.ServerBase, id, attachMode, 1*time.Second)
		if snapErr != nil {
			return nil, snapErr
		}
		resp.WSOutput = out
		if strings.TrimSpace(out) != "" {
			resp.SnapshotCount = 1
		}
	}

	for i := resp.SnapshotCount; i < n; i++ {
		out, snapErr := wsAttachOnly(t, req.ServerBase, id, attachMode, 1*time.Second)
		if snapErr != nil {
			return nil, fmt.Errorf("snapshot %d/%d: %w", i+1, n, snapErr)
		}
		if strings.TrimSpace(out) != "" {
			resp.SnapshotCount++
			resp.WSOutput = out
		}
		// Close happens inside wsAttachOnly; brief pause between short-lived attaches.
		time.Sleep(50 * time.Millisecond)
	}

	// Allow any stopChild race to settle if attach mode incorrectly claimed writer.
	time.Sleep(300 * time.Millisecond)

	sessions, err := listSessions(t, req.ServerBase)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.Sessions = sessions
	resp.ProcessAlive = processAlive(pid)
	resp.SessionListed = sessionInList(sessions, id)
	resp.SessionCount = len(sessions)
	if resp.ProcessAlive {
		resp.RunningProcessCount = 1
	}
	t.Cleanup(func() {
		if processAlive(pid) {
			_ = exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
		}
		_ = deleteSessionREST(req.ServerBase, id)
	})
	return resp, nil
}

// runWriterCloseLifecycle creates a long-lived sleep PTY, attaches as writer,
// closes the WS with req.WSCloseCode, then reports whether the child still runs
// and whether the session remains listed.
//
// Close code 1000 (normal) currently detaches without killing the child — that
// is the PTY leak path when clients churn terminals. Expected correct behavior
// after fix: child is reaped (ProcessAlive=false) so the OS PTY is released.
// Close code 4000 must remove the session and kill the child.
func runWriterCloseLifecycle(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	code := req.WSCloseCode
	if code == 0 {
		code = 1000
	}
	marker := fmt.Sprintf("ptywrap-leak-%d-%d", os.Getpid(), time.Now().UnixNano())
	// Unique sleep argv so we can resolve the child PID without /proc.
	cmd := []string{"sleep", "3600"}
	beforeSleep := snapshotSleep3600PIDs()
	id, err := createSessionREST(t, req.ServerBase, cmd, "", marker)
	if err != nil {
		return nil, err
	}
	pid, err := findNewSleepPID(t, beforeSleep)
	if err != nil {
		return nil, fmt.Errorf("session %s: %w", id, err)
	}
	if err := wsCloseWriter(t, req.ServerBase, id, code); err != nil {
		return nil, err
	}
	time.Sleep(400 * time.Millisecond)

	sessions, err := listSessions(t, req.ServerBase)
	if err != nil {
		return nil, err
	}
	resp.SessionID = id
	resp.Sessions = sessions
	resp.ProcessAlive = processAlive(pid)
	resp.SessionListed = sessionInList(sessions, id)
	resp.SessionCount = len(sessions)
	if resp.ProcessAlive {
		resp.RunningProcessCount = 1
	}
	// Best-effort cleanup so parallel doctests do not accumulate sleeps if the
	// assertion fails early.
	t.Cleanup(func() {
		if processAlive(pid) {
			_ = exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
		}
		_ = deleteSessionREST(req.ServerBase, id)
	})
	return resp, nil
}

// runMultiCreateOnConnectOrphan models LocalTerminal / StrictMode churn:
// N times WS connect without session_id (creates a shell), then normal close
// (code 1000). Expected correct behavior: no orphan shell processes remain
// (RunningProcessCount == 0). Buggy behavior leaves one live bash per connect.
func runMultiCreateOnConnectOrphan(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	n := req.RepeatCount
	if n <= 0 {
		n = 5
	}
	var pids []int
	var ids []string
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("orphan-%d-%d", os.Getpid(), i)
		before := snapshotPtywrapBashPIDs()
		id, err := wsCreateOnConnectClose(t, req.ServerBase, name, 1000)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
		time.Sleep(150 * time.Millisecond)
		after := snapshotPtywrapBashPIDs()
		pid := newestPIDNotIn(after, before)
		if pid > 0 {
			pids = append(pids, pid)
		}
	}
	time.Sleep(400 * time.Millisecond)

	alive := 0
	for _, pid := range pids {
		if processAlive(pid) {
			alive++
		}
	}
	sessions, err := listSessions(t, req.ServerBase)
	if err != nil {
		return nil, err
	}
	resp.Sessions = sessions
	resp.SessionCount = len(sessions)
	resp.RunningProcessCount = alive
	resp.ProcessAlive = alive > 0
	if len(ids) > 0 {
		resp.SessionID = ids[len(ids)-1]
		resp.SessionListed = sessionInList(sessions, resp.SessionID)
	}
	t.Cleanup(func() {
		for _, pid := range pids {
			if processAlive(pid) {
				_ = exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
			}
		}
		for _, id := range ids {
			_ = deleteSessionREST(req.ServerBase, id)
		}
	})
	return resp, nil
}

// runShellExitWSCloseCode attaches as writer while a short-lived shell is still
// running, waits for the shell to exit (server-side <-s.done), and records the
// WebSocket close code the client observes. Correct behavior: 1000 Normal
// Closure. Buggy behavior: bare conn.Close() → 1006 / unexpected EOF.
func runShellExitWSCloseCode(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	// Sleep briefly so attach claims writer before exit; then process ends.
	cmd := req.Command
	if len(cmd) == 0 {
		cmd = []string{"sh", "-c", "sleep 1"}
	}
	id, err := createSessionREST(t, req.ServerBase, cmd, "", "shell-exit-ws")
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() {
		_ = deleteSessionREST(req.ServerBase, id)
	})

	conn := dialWS(t, req.ServerBase, "session_id="+url.QueryEscape(id))
	defer conn.Close()
	recv := startWSReader(conn)
	drainWSHandshake(recv, 500*time.Millisecond)

	code, out, waitErr := waitWSServerClose(recv, 15*time.Second)
	if waitErr != nil {
		return nil, waitErr
	}
	resp.SessionID = id
	resp.CloseCode = code
	resp.WSOutput = out

	// Give list endpoint a moment to reflect exited status.
	time.Sleep(200 * time.Millisecond)
	sessions, err := listSessions(t, req.ServerBase)
	if err != nil {
		return nil, err
	}
	resp.Sessions = sessions
	resp.SessionListed = sessionInList(sessions, id)
	resp.SessionCount = len(sessions)
	resp.SessionStatus = sessionStatus(sessions, id)
	return resp, nil
}

// runShellExitAttachWait uses ptywrap/client AttachWithIO(Wait=true) against a
// short-lived session. After shell exit, Attach must return nil (normalize
// treats close 1000 as success). Buggy bare close yields
// "terminal closed: unexpected EOF" / unexpected EOF.
func runShellExitAttachWait(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	cmd := req.Command
	if len(cmd) == 0 {
		cmd = []string{"sh", "-c", "sleep 1"}
	}
	id, err := createSessionREST(t, req.ServerBase, cmd, "", "shell-exit-attach")
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() {
		_ = deleteSessionREST(req.ServerBase, id)
	})

	c := ptyclient.NewClient(req.ServerBase)
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer stdinR.Close()
	// Keep write end open so stdin does not EOF immediately and race the bridge.
	t.Cleanup(func() { _ = stdinW.Close() })

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer stdoutW.Close()
	// Drain stdout so the pipe does not block the bridge.
	go io.Copy(io.Discard, stdoutR)

	opts := ptyclient.ConnectOptions{
		SessionID:    id,
		Wait:         true,
		SkipTTYCheck: true,
	}
	_, attachErr := ptyclient.AttachWithIO(c, opts, stdinR, stdoutW, stdoutW)
	if attachErr != nil {
		resp.AttachErr = attachErr.Error()
	}
	resp.SessionID = id

	sessions, listErr := listSessions(t, req.ServerBase)
	if listErr != nil {
		return nil, listErr
	}
	resp.Sessions = sessions
	resp.SessionListed = sessionInList(sessions, id)
	resp.SessionCount = len(sessions)
	resp.SessionStatus = sessionStatus(sessions, id)
	return resp, nil
}

// runShellExitMarkerWithoutClose proves Attach Wait ends on the session-exit
// text marker alone: mock server sends session_id + "[Terminal exited]" and
// never closes the socket. Correct client returns nil within a short deadline
// (does not hang waiting for close 1000 / 1006).
func runShellExitMarkerWithoutClose(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	_ = req

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Keep conn open until test ends; do not send a close frame.
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"session_id","session_id":"mock-exit-marker"}`))
		time.Sleep(50 * time.Millisecond)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\n[Terminal exited]"))
		// Block until client disconnects so the socket stays open without close.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				_ = conn.Close()
				return
			}
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := ptyclient.NewClient(srv.URL)
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer stdinR.Close()
	t.Cleanup(func() { _ = stdinW.Close() })

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer stdoutW.Close()
	go io.Copy(io.Discard, stdoutR)

	done := make(chan error, 1)
	go func() {
		_, attachErr := ptyclient.AttachWithIO(c, ptyclient.ConnectOptions{
			// Known id skips strict handshake timeout if needed; mock always sends session_id.
			SessionID:    "mock-exit-marker",
			Wait:         true,
			SkipTTYCheck: true,
		}, stdinR, stdoutW, stdoutW)
		done <- attachErr
	}()

	select {
	case attachErr := <-done:
		if attachErr != nil {
			resp.AttachErr = attachErr.Error()
		}
	case <-time.After(3 * time.Second):
		resp.AttachErr = "timeout_hang: attach did not return after exit marker without close"
	}
	resp.SessionID = "mock-exit-marker"
	return resp, nil
}

// runShellExitHardDropWithoutMarker proves Attach Wait still errors when the
// peer tears down the socket without sending "[Terminal exited]" and without
// a normal close frame. Negative control: do not silence 1006/EOF globally.
func runShellExitHardDropWithoutMarker(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	_ = req

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"session_id","session_id":"mock-hard-drop"}`))
		time.Sleep(50 * time.Millisecond)
		// Bare tear-down: no exit marker, no FormatCloseMessage(1000).
		_ = conn.Close()
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := ptyclient.NewClient(srv.URL)
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer stdinR.Close()
	t.Cleanup(func() { _ = stdinW.Close() })

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer stdoutW.Close()
	go io.Copy(io.Discard, stdoutR)

	done := make(chan error, 1)
	go func() {
		_, attachErr := ptyclient.AttachWithIO(c, ptyclient.ConnectOptions{
			SessionID:    "mock-hard-drop",
			Wait:         true,
			SkipTTYCheck: true,
		}, stdinR, stdoutW, stdoutW)
		done <- attachErr
	}()

	select {
	case attachErr := <-done:
		if attachErr != nil {
			resp.AttachErr = attachErr.Error()
		}
	case <-time.After(3 * time.Second):
		resp.AttachErr = "timeout_hang: attach did not return after hard drop without marker"
	}
	resp.SessionID = "mock-hard-drop"
	return resp, nil
}

// waitWSServerClose drains frames until the reader reports an error (typically
// a server-initiated WebSocket close) and returns the close code.
func waitWSServerClose(recv <-chan wsFrame, timeout time.Duration) (closeCode int, out string, err error) {
	deadline := time.After(timeout)
	var buf strings.Builder
	for {
		select {
		case <-deadline:
			return 0, buf.String(), fmt.Errorf("timeout waiting for server-initiated WebSocket close")
		case f, ok := <-recv:
			if !ok {
				if buf.Len() > 0 {
					// Reader ended without an explicit close error; treat as abnormal.
					return websocket.CloseAbnormalClosure, buf.String(), nil
				}
				return 0, "", fmt.Errorf("websocket reader closed before server close")
			}
			if f.err != nil {
				code := websocket.CloseAbnormalClosure
				if ce, ok := f.err.(*websocket.CloseError); ok {
					code = ce.Code
				} else if strings.Contains(f.err.Error(), "unexpected EOF") {
					code = websocket.CloseAbnormalClosure
				}
				return code, buf.String(), nil
			}
			if f.mt == websocket.TextMessage && isControlJSON(f.msg) {
				continue
			}
			if f.mt == websocket.BinaryMessage || f.mt == websocket.TextMessage {
				appendFilteredWS(&buf, f.msg)
			}
		}
	}
}

func sessionStatus(sessions []ptywrap.SessionInfo, id string) string {
	for _, s := range sessions {
		if s.ID == id {
			return s.Status
		}
	}
	return ""
}

func findNewSleepPID(t *testing.T, before map[int]struct{}) (int, error) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		after := snapshotSleep3600PIDs()
		if pid := newestPIDNotIn(after, before); pid > 0 && processAlive(pid) {
			return pid, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return 0, fmt.Errorf("could not resolve new sleep 3600 child")
}

func snapshotSleep3600PIDs() map[int]struct{} {
	out, err := exec.Command("pgrep", "-f", "sleep 3600").Output()
	m := make(map[int]struct{})
	if err != nil {
		return m
	}
	for _, f := range strings.Fields(string(out)) {
		pid, err := strconv.Atoi(f)
		if err == nil && pid > 0 {
			m[pid] = struct{}{}
		}
	}
	return m
}

func wsCloseWriter(t *testing.T, base, sessionID string, closeCode int) error {
	t.Helper()
	conn := dialWS(t, base, "session_id="+url.QueryEscape(sessionID))
	recv := startWSReader(conn)
	drainWSHandshake(recv, 500*time.Millisecond)
	msg := websocket.FormatCloseMessage(closeCode, "")
	_ = conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))
	_ = conn.Close()
	return nil
}

func wsCreateOnConnectClose(t *testing.T, base, name string, closeCode int) (string, error) {
	t.Helper()
	q := url.Values{}
	q.Set("name", name)
	conn := dialWS(t, base, q.Encode())
	id := readSessionIDMessage(t, conn)
	msg := websocket.FormatCloseMessage(closeCode, "")
	_ = conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))
	_ = conn.Close()
	return id, nil
}

func deleteSessionREST(base, id string) error {
	req, err := http.NewRequest(http.MethodDelete, base+"/api/terminal/sessions?id="+url.QueryEscape(id), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return exec.Command("kill", "-0", strconv.Itoa(pid)).Run() == nil
}

func sessionInList(sessions []ptywrap.SessionInfo, id string) bool {
	for _, s := range sessions {
		if s.ID == id {
			return true
		}
	}
	return false
}

func snapshotPtywrapBashPIDs() map[int]struct{} {
	out, err := exec.Command("ps", "-axo", "pid=,command=").Output()
	if err != nil {
		return map[int]struct{}{}
	}
	m := make(map[int]struct{})
	for _, line := range strings.Split(string(out), "\n") {
		// Default shell spawn uses --rcfile .../.ptywrap-bashrc
		if !strings.Contains(line, "ptywrap-bashrc") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err == nil && pid > 0 {
			m[pid] = struct{}{}
		}
	}
	return m
}

func newestPIDNotIn(after, before map[int]struct{}) int {
	newest := 0
	for pid := range after {
		if _, ok := before[pid]; ok {
			continue
		}
		if pid > newest {
			newest = pid
		}
	}
	return newest
}

func createSessionREST(t *testing.T, base string, command []string, cwd, name string) (string, error) {
	body, status, err := postCreateSession(t, base, command, cwd, name)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return "", fmt.Errorf("create status %d: %v", status, body)
	}
	id, _ := body["id"].(string)
	if id == "" {
		return "", fmt.Errorf("missing id in create response: %v", body)
	}
	return id, nil
}

func postCreateSession(t *testing.T, base string, command []string, cwd, name string) (map[string]interface{}, int, error) {
	payload := map[string]interface{}{}
	if len(command) > 0 {
		payload["command"] = command[0]
		if len(command) > 1 {
			payload["args"] = command[1:]
		}
	}
	if cwd != "" {
		payload["cwd"] = cwd
	}
	if name != "" {
		payload["name"] = name
	}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(base+"/api/terminal/sessions", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	var out map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&out)
	return out, resp.StatusCode, nil
}

func patchRenameSession(t *testing.T, base, id, name string) error {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest(http.MethodPatch, base+"/api/terminal/sessions/"+url.PathEscape(id), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rename status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func listSessions(t *testing.T, base string) ([]ptywrap.SessionInfo, error) {
	resp, err := http.Get(base + "/api/terminal/sessions?page_size=100")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var page struct {
		Sessions []ptywrap.SessionInfo `json:"sessions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}
	return page.Sessions, nil
}

func wsURL(base, query string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path = "/api/terminal"
	u.RawQuery = query
	return u.String(), nil
}

func dialWS(t *testing.T, base, query string) *websocket.Conn {
	t.Helper()
	ws, err := wsURL(base, query)
	if err != nil {
		t.Fatal(err)
	}
	conn, _, err := websocket.DefaultDialer.Dial(ws, nil)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func readSessionIDMessage(t *testing.T, conn *websocket.Conn) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		var m struct {
			Type      string `json:"type"`
			SessionID string `json:"session_id"`
		}
		if json.Unmarshal(msg, &m) == nil && m.Type == "session_id" && m.SessionID != "" {
			return m.SessionID
		}
	}
	t.Fatal("timeout waiting for session_id message")
	return ""
}

func wsCreateOnConnect(t *testing.T, base, name, cwd string) (string, string, error) {
	q := url.Values{}
	q.Set("name", name)
	q.Set("cwd", cwd)
	conn := dialWS(t, base, q.Encode())
	defer conn.Close()
	id := readSessionIDMessage(t, conn)
	return id, "", nil
}

func wsCollectOutput(t *testing.T, base, sessionID string, wait time.Duration) (string, error) {
	conn := dialWS(t, base, "session_id="+url.QueryEscape(sessionID))
	defer conn.Close()
	return collectWSBinary(conn, wait)
}

func wsAttachOnly(t *testing.T, base, sessionID, attachMode string, wait time.Duration) (string, error) {
	q := url.Values{"session_id": {sessionID}}
	if attachMode != "" {
		q.Set("attach_mode", attachMode)
	}
	conn := dialWS(t, base, q.Encode())
	defer conn.Close()
	return collectWSBinary(conn, wait)
}

func wsRunCommand(t *testing.T, base, sessionID, input string) (string, error) {
	conn := dialWS(t, base, "session_id="+url.QueryEscape(sessionID))
	defer conn.Close()
	recv := startWSReader(conn)
	drainWSHandshake(recv, 2*time.Second)
	time.Sleep(300 * time.Millisecond)
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte(input)); err != nil {
		return "", err
	}
	// Match actual TTY probe result (tty=0/1), not the echoed python source
	// which also contains the substring "tty=". Otherwise fall back to the
	// full input payload (cat echo round-trip).
	want := strings.TrimSpace(input)
	return collectWSUntil(recv, 10*time.Second, func(s string) bool {
		if strings.Contains(s, "tty=0") || strings.Contains(s, "tty=1") {
			return true
		}
		// Avoid matching the command echo for the python TTY probe.
		if strings.Contains(want, "tty=") {
			return false
		}
		return want != "" && strings.Contains(s, want)
	})
}

func wsResizeAndCollect(t *testing.T, base, sessionID string, cols, rows int, wait time.Duration) (string, error) {
	conn := dialWS(t, base, "session_id="+url.QueryEscape(sessionID))
	defer conn.Close()
	recv := startWSReader(conn)
	drainWSHandshake(recv, 2*time.Second)
	resizeMsg := map[string]interface{}{"type": "resize", "cols": cols, "rows": rows}
	msg, _ := json.Marshal(resizeMsg)
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return "", err
	}
	time.Sleep(300 * time.Millisecond)
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte("stty size\n")); err != nil {
		return "", err
	}
	return collectWSUntil(recv, wait, func(s string) bool {
		c, r := parseSttySize(s)
		return c == cols && r == rows
	})
}

func wsWaitDone(t *testing.T, base, sessionID string, timeout time.Duration) (string, error) {
	conn := dialWS(t, base, "session_id="+url.QueryEscape(sessionID))
	defer conn.Close()
	recv := startWSReader(conn)
	return collectWSUntil(recv, timeout, func(s string) bool {
		return strings.Contains(s, "[Terminal exited]")
	})
}

// wsFrame is one message from a single background ReadMessage loop.
// gorilla/websocket forbids SetReadDeadline timeouts for polling: after a read
// times out the connection is permanently unusable. Callers must use one
// reader goroutine and select on wall-clock deadlines instead.
type wsFrame struct {
	mt  int
	msg []byte
	err error
}

func startWSReader(conn *websocket.Conn) <-chan wsFrame {
	ch := make(chan wsFrame, 64)
	go func() {
		defer close(ch)
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				ch <- wsFrame{err: err}
				return
			}
			// Copy payload; gorilla may reuse the buffer.
			payload := append([]byte(nil), msg...)
			ch <- wsFrame{mt: mt, msg: payload}
		}
	}()
	return ch
}

func drainWSHandshake(recv <-chan wsFrame, timeout time.Duration) {
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			return
		case f, ok := <-recv:
			if !ok || f.err != nil {
				return
			}
			if f.mt == websocket.TextMessage && isControlJSON(f.msg) {
				continue
			}
			if f.mt == websocket.BinaryMessage || f.mt == websocket.TextMessage {
				// First non-control payload (e.g. scrollback / prompt) ends drain.
				return
			}
		}
	}
}

func collectWSBinary(conn *websocket.Conn, wait time.Duration) (string, error) {
	return collectWSUntil(startWSReader(conn), wait, nil)
}

func collectWSUntil(recv <-chan wsFrame, wait time.Duration, done func(string) bool) (string, error) {
	deadline := time.After(wait)
	var buf strings.Builder
	for {
		if done != nil && done(buf.String()) {
			return buf.String(), nil
		}
		select {
		case <-deadline:
			return buf.String(), nil
		case f, ok := <-recv:
			if !ok {
				if buf.Len() > 0 {
					return buf.String(), nil
				}
				return "", fmt.Errorf("websocket reader closed")
			}
			if f.err != nil {
				if buf.Len() > 0 {
					return buf.String(), nil
				}
				return "", f.err
			}
			if f.mt == websocket.TextMessage && isControlJSON(f.msg) {
				continue
			}
			if f.mt == websocket.BinaryMessage || f.mt == websocket.TextMessage {
				appendFilteredWS(&buf, f.msg)
				if done != nil && done(buf.String()) {
					return buf.String(), nil
				}
			}
		}
	}
}

func appendFilteredWS(buf *strings.Builder, msg []byte) {
	s := string(msg)
	for {
		idx := strings.Index(s, `{"type":"session_id"`)
		if idx < 0 {
			buf.WriteString(s)
			return
		}
		buf.WriteString(s[:idx])
		end := strings.Index(s[idx:], "}")
		if end < 0 {
			return
		}
		s = s[idx+end+1:]
	}
}

func isControlJSON(msg []byte) bool {
	s := strings.TrimSpace(string(msg))
	if strings.Contains(s, `"type":"session_id"`) || strings.Contains(s, `"type": "session_id"`) {
		return true
	}
	var m struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(msg, &m) != nil {
		return false
	}
	switch m.Type {
	case "session_id", "error":
		return true
	default:
		return m.Type != ""
	}
}

var sttySizeRE = regexp.MustCompile(`(\d+)\s+(\d+)`)

func parseSttySize(out string) (int, int) {
	matches := sttySizeRE.FindAllStringSubmatch(out, -1)
	if len(matches) == 0 {
		return 0, 0
	}
	last := matches[len(matches)-1]
	var rows, cols int
	fmt.Sscanf(last[1], "%d", &rows)
	fmt.Sscanf(last[2], "%d", &cols)
	if rows <= 0 || cols <= 0 {
		return 0, 0
	}
	return cols, rows
}