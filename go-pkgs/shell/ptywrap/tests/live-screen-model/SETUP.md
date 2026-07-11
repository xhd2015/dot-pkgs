# Scenario

**Feature**: attach_mode=snapshot exports a persistent live screen model

```
# PTY child paints sticky chrome once, then dirty-only frames
fixture TUI -> PTY output chunks
  -> session.screen.Write(chunk)   # live VT (source of truth)
  -> session.scrollback append+trim

# Snapshot does not cold-replay truncated ring
attach_mode=snapshot
  -> exportCells(session.screen)
  -> WS binary frame contains STICKY_FOOTER / STICKY_PROMPT
```

## Preconditions

1. Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` importable.
2. `python3` available (ANSI fixture driver).
3. Production scrollback cap is 256 KiB (`maxScrollback`); pressure scenarios
   emit past that so early sticky paint leaves the ring under cold replay.
4. Session cache: none required — each leaf starts its own httptest server via
   root `Setup` (mirrors `snapshot-attach`).

## Steps

1. Root `Setup` starts ephemeral HTTP test server with `ptywrap.RegisterAPI`.
2. Leaf sets `Phase` + stress parameters (dirty iters / pressure / resize / N).
3. `Run` launches fixture TUI, snapshots, reports markers + liveness.

## Context

Cold scrollback replay (`renderScreenSnapshot(scrollback, cols, rows)` on a
fresh VT) loses sticky chrome once dirty-region frames dominate the ring.
Live VT keeps cells from the first sticky paint until overwritten.

Shared HTTP/WS helpers below are used by root `Run` only (not redefined by leaves).

```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

func Setup(t *testing.T, req *Request) error {
	base, cleanup := startTestServer(t)
	t.Cleanup(cleanup)
	req.ServerBase = base
	return nil
}

func startTestServer(t *testing.T) (base string, cleanup func()) {
	t.Helper()
	mux := http.NewServeMux()
	ptywrap.RegisterAPI(mux)
	srv := httptest.NewServer(mux)
	return srv.URL, srv.Close
}

func createSessionREST(t *testing.T, base string, command []string, cwd, name string) (string, error) {
	t.Helper()
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
	t.Helper()
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
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out, resp.StatusCode, nil
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

func listSessions(t *testing.T, base string) ([]ptywrap.SessionInfo, error) {
	t.Helper()
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

func sessionInList(sessions []ptywrap.SessionInfo, id string) bool {
	for _, s := range sessions {
		if s.ID == id {
			return true
		}
	}
	return false
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return exec.Command("kill", "-0", strconv.Itoa(pid)).Run() == nil
}

func killPID(pid int) error {
	return exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
}

func findPIDByToken(t *testing.T, token string) (int, error) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		out, err := exec.Command("pgrep", "-f", token).Output()
		if err == nil {
			fields := strings.Fields(string(out))
			// Prefer highest PID (newest) among matches.
			best := 0
			for _, f := range fields {
				pid, err := strconv.Atoi(f)
				if err == nil && pid > best && processAlive(pid) {
					best = pid
				}
			}
			if best > 0 {
				return best, nil
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return 0, fmt.Errorf("could not resolve child PID for token %q", token)
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
			payload := append([]byte(nil), msg...)
			ch <- wsFrame{mt: mt, msg: payload}
		}
	}()
	return ch
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
	case "session_id", "error", "attach_role":
		return true
	default:
		return m.Type != ""
	}
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
				buf.Write(f.msg)
				if done != nil && done(buf.String()) {
					return buf.String(), nil
				}
			}
		}
	}
}

// wsAttachSnapshot opens attach_mode=snapshot (or given mode), collects the
// one-shot frame, and closes the socket.
func wsAttachSnapshot(t *testing.T, base, sessionID, attachMode string, wait time.Duration) (string, error) {
	t.Helper()
	q := url.Values{"session_id": {sessionID}}
	if attachMode != "" {
		q.Set("attach_mode", attachMode)
	}
	conn := dialWS(t, base, q.Encode())
	defer conn.Close()
	return collectWSUntil(startWSReader(conn), wait, nil)
}

// wsResizeOnly attaches as writer long enough to send a resize control message.
func wsResizeOnly(t *testing.T, base, sessionID string, cols, rows int) error {
	t.Helper()
	conn := dialWS(t, base, "session_id="+url.QueryEscape(sessionID))
	defer conn.Close()
	recv := startWSReader(conn)
	// Drain handshake / initial frame briefly.
	_, _ = collectWSUntil(recv, 500*time.Millisecond, nil)
	msg, _ := json.Marshal(map[string]interface{}{
		"type": "resize",
		"cols": cols,
		"rows": rows,
	})
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return err
	}
	// Keep writer briefly so resize is processed; then normal close.
	time.Sleep(200 * time.Millisecond)
	_ = conn.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second))
	return nil
}
```
