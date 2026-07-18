package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

type controlMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

type serverMessage struct {
	Type      string `json:"type"`
	Message   string `json:"message,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

type wsWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsWriter) writeMessage(messageType int, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(messageType, data)
}

func (w *wsWriter) writeJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return w.writeMessage(websocket.TextMessage, data)
}

func (w *wsWriter) close(code int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	msg := websocket.FormatCloseMessage(code, "")
	return w.conn.WriteControl(websocket.CloseMessage, msg, time.Time{})
}

// Attach connects to a terminal session. When Wait is true, requires an
// interactive terminal on stdin and stdout and blocks until disconnect.
func Attach(c *Client, opts ConnectOptions) (AttachResult, error) {
	if opts.Wait {
		return AttachWithIO(c, opts, os.Stdin, os.Stdout, os.Stderr)
	}
	opts.SkipTTYCheck = true
	return AttachWithIO(c, opts, strings.NewReader(""), io.Discard, io.Discard)
}

// AttachWithIO bridges local IO to the daemon WebSocket.
func AttachWithIO(c *Client, opts ConnectOptions, stdin io.Reader, stdout io.Writer, stderr io.Writer) (AttachResult, error) {
	_ = stderr
	if !opts.SkipTTYCheck {
		if forceNonTTYTest() || !isInteractiveIO(stdin, stdout) {
			return AttachResult{}, fmt.Errorf("attach requires an interactive terminal on stdin/stdout")
		}
	}

	wsURL, err := terminalWebSocketURL(c.BaseURL, opts)
	if err != nil {
		return AttachResult{}, err
	}

	header := http.Header{}
	token := opts.AuthToken
	if token == "" {
		token = c.AuthToken
	}
	if token != "" {
		header.Set("Authorization", "Bearer "+token)
	}

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return AttachResult{}, terminalDialError(err, resp)
	}

	result, err := readSessionID(conn, opts.SessionID)
	if err != nil {
		conn.Close()
		return AttachResult{}, err
	}

	if !opts.Wait {
		go func() {
			defer conn.Close()
			runBridge(conn, stdin, stdout)
		}()
		return result, nil
	}

	defer conn.Close()
	runErr := runBridge(conn, stdin, stdout)
	return result, runErr
}

func runBridge(conn *websocket.Conn, stdin io.Reader, stdout io.Writer) error {
	writer := &wsWriter{conn: conn}

	stdinFile, hasTTY := stdin.(*os.File)
	var oldState *term.State
	if hasTTY {
		if state, err := term.MakeRaw(int(stdinFile.Fd())); err == nil {
			oldState = state
			defer term.Restore(int(stdinFile.Fd()), oldState)
		}
	}

	if stdoutFile, ok := stdout.(*os.File); ok {
		_ = sendTerminalResize(writer, stdoutFile)
		if hasTTY {
			sigWinch := make(chan os.Signal, 1)
			signal.Notify(sigWinch, syscall.SIGWINCH)
			defer signal.Stop(sigWinch)
			go func() {
				for range sigWinch {
					_ = sendTerminalResize(writer, stdoutFile)
				}
			}()
		}
	}

	readerErrCh := make(chan error, 1)
	go func() {
		readerErrCh <- readTerminalOutput(conn, stdout)
	}()

	stdinErrCh := make(chan error, 1)
	go func() {
		stdinErrCh <- forwardTerminalInput(writer, stdin)
	}()

	var runErr error
	select {
	case err := <-readerErrCh:
		// Session-end marker returns nil from the reader; treat that as clean
		// success even if a later close would have been 1006.
		runErr = normalizeTerminalReadError(err)
	case err := <-stdinErrCh:
		if err != nil && err != io.EOF {
			runErr = err
		}
	}

	_ = writer.close(websocket.CloseNormalClosure)
	return runErr
}

// readSessionID reads the server's session_id handshake frame. knownSessionID
// is the SessionID the caller is attaching to ("" for create/new mode).
//
// In attach mode (knownSessionID != "") the client is lenient: a pre-refactor
// daemon reattaches by serving the session immediately (writing a binary
// scrollback/output frame) without echoing a session_id frame. Rather than
// timing out against such a server, the client proceeds with the known
// SessionID. The consumed binary frame is best-effort (a scrollback snapshot).
func readSessionID(conn *websocket.Conn, knownSessionID string) (AttachResult, error) {
	// Set a single read deadline for the whole handshake window. gorilla
	// permanently poisons the connection's readErr on a deadline timeout
	// (hideTempErr wraps it as a non-temporary netError), so the connection
	// cannot be reused for reads after the first timeout. Re-arming the
	// deadline each iteration and continuing therefore spins ReadMessage
	// against the cached readErr until gorilla panics with
	// "repeated read on failed websocket connection". Use one deadline and
	// surface any error immediately instead.
	deadline := time.Now().Add(10 * time.Second)
	_ = conn.SetReadDeadline(deadline)
	for time.Now().Before(deadline) {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				// No session_id arrived. In attach mode, a silent server is
				// the pre-refactor reattach behavior — proceed with the known
				// SessionID instead of timing out.
				if knownSessionID != "" {
					_ = conn.SetReadDeadline(time.Time{})
					return AttachResult{SessionID: knownSessionID}, nil
				}
				return AttachResult{}, fmt.Errorf("timeout waiting for session_id message")
			}
			return AttachResult{}, err
		}
		if msgType == websocket.TextMessage {
			handled, sessionID, err := parseServerMessage(data)
			if err != nil {
				return AttachResult{}, err
			}
			if handled && sessionID != "" {
				_ = conn.SetReadDeadline(time.Time{})
				return AttachResult{SessionID: sessionID}, nil
			}
		} else {
			// Binary frame: the server has started serving without a
			// session_id handshake (pre-refactor reattach). In attach mode,
			// proceed with the known SessionID.
			if knownSessionID != "" {
				_ = conn.SetReadDeadline(time.Time{})
				return AttachResult{SessionID: knownSessionID}, nil
			}
		}
	}
	if knownSessionID != "" {
		_ = conn.SetReadDeadline(time.Time{})
		return AttachResult{SessionID: knownSessionID}, nil
	}
	return AttachResult{}, fmt.Errorf("timeout waiting for session_id message")
}

func parseServerMessage(data []byte) (handled bool, sessionID string, err error) {
	var msg serverMessage
	if json.Unmarshal(data, &msg) != nil || msg.Type == "" {
		return false, "", nil
	}
	switch msg.Type {
	case "session_id":
		return true, msg.SessionID, nil
	case "error":
		if msg.Message == "" {
			return true, "", fmt.Errorf("remote terminal error")
		}
		return true, "", fmt.Errorf("%s", msg.Message)
	default:
		return true, "", nil
	}
}

func terminalWebSocketURL(base string, opts ConnectOptions) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid server url %q: %w", base, err)
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		return "", fmt.Errorf("unsupported server scheme %q", u.Scheme)
	}
	u.Path = "/api/terminal"
	q := u.Query()
	if strings.TrimSpace(opts.SessionID) != "" {
		q.Set("session_id", opts.SessionID)
	}
	if opts.AttachSnapshot {
		q.Set("attach_mode", "screen")
	}
	if strings.TrimSpace(opts.Name) != "" {
		q.Set("name", opts.Name)
	}
	if strings.TrimSpace(opts.Cwd) != "" {
		q.Set("cwd", opts.Cwd)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func sendTerminalResize(writer *wsWriter, stdout *os.File) error {
	cols, rows, err := term.GetSize(int(stdout.Fd()))
	if err != nil {
		return nil
	}
	return writer.writeJSON(controlMessage{Type: "resize", Cols: cols, Rows: rows})
}

func forwardTerminalInput(writer *wsWriter, stdin io.Reader) error {
	buf := make([]byte, 4096)
	for {
		n, err := stdin.Read(buf)
		if n > 0 {
			if writeErr := writer.writeMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			return err
		}
	}
}

func readTerminalOutput(conn *websocket.Conn, stdout io.Writer) error {
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		switch msgType {
		case websocket.BinaryMessage:
			if _, err := stdout.Write(data); err != nil {
				return err
			}
		case websocket.TextMessage:
			// Session exited: end the attach Wait on status, not only on socket
			// death. Waiting for a close frame hangs when the peer tears down
			// without a normal close (1006) or when the close is lost.
			if isTerminalExitMarker(data) {
				return nil
			}
			handled, _, err := parseServerMessage(data)
			if err != nil {
				return err
			}
			if !handled {
				if _, err := stdout.Write(data); err != nil {
					return err
				}
			}
		}
	}
}

func isTerminalExitMarker(data []byte) bool {
	return strings.Contains(string(data), "[Terminal exited]")
}

func normalizeTerminalReadError(err error) error {
	if err == nil {
		return nil
	}
	if ce, ok := err.(*websocket.CloseError); ok {
		switch ce.Code {
		case websocket.CloseNormalClosure, websocket.CloseGoingAway, 4000:
			return nil
		}
		return fmt.Errorf("terminal closed: %s", ce.Text)
	}
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		return nil
	}
	if strings.Contains(err.Error(), "use of closed network connection") {
		return nil
	}
	return err
}

func terminalDialError(err error, resp *http.Response) error {
	if resp == nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		return fmt.Errorf("terminal connect failed: %s", resp.Status)
	}
	return fmt.Errorf("terminal connect failed: %s: %s", resp.Status, snippet)
}

func forceNonTTYTest() bool {
	return os.Getenv("PTYWRAP_CLIENT_TEST_STDIN_FD") == "pipe"
}

func isInteractiveIO(stdin io.Reader, stdout io.Writer) bool {
	in, okIn := stdin.(*os.File)
	out, okOut := stdout.(*os.File)
	if !okIn || !okOut {
		return false
	}
	if !term.IsTerminal(int(in.Fd())) || !term.IsTerminal(int(out.Fd())) {
		return false
	}
	if st, err := in.Stat(); err == nil && st.Mode()&os.ModeNamedPipe != 0 {
		return false
	}
	if st, err := out.Stat(); err == nil && st.Mode()&os.ModeNamedPipe != 0 {
		return false
	}
	return true
}

func isTerminalReader(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// OpenTestPTY opens a pseudo-terminal master for doctest harnesses.
func OpenTestPTY() (*os.File, error) {
	ptmx, _, err := pty.Open()
	return ptmx, err
}