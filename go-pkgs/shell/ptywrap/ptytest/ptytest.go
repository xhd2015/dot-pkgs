package ptytest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
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
	id, err := createSessionREST(t, req.ServerBase, []string{"sh", "-c", "sleep 30"}, "", "")
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
	drainWSHandshake(conn, 2*time.Second)
	time.Sleep(300 * time.Millisecond)
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte(input)); err != nil {
		return "", err
	}
	return collectWSUntil(conn, 10*time.Second, func(s string) bool {
		return strings.Contains(s, "tty=")
	})
}

func wsResizeAndCollect(t *testing.T, base, sessionID string, cols, rows int, wait time.Duration) (string, error) {
	conn := dialWS(t, base, "session_id="+url.QueryEscape(sessionID))
	defer conn.Close()
	drainWSHandshake(conn, 2*time.Second)
	resizeMsg := map[string]interface{}{"type": "resize", "cols": cols, "rows": rows}
	msg, _ := json.Marshal(resizeMsg)
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return "", err
	}
	time.Sleep(300 * time.Millisecond)
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte("stty size\n")); err != nil {
		return "", err
	}
	return collectWSUntil(conn, wait, func(s string) bool {
		c, r := parseSttySize(s)
		return c == cols && r == rows
	})
}

func wsWaitDone(t *testing.T, base, sessionID string, timeout time.Duration) (string, error) {
	conn := dialWS(t, base, "session_id="+url.QueryEscape(sessionID))
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var buf strings.Builder
	for {
		select {
		case <-ctx.Done():
			return buf.String(), ctx.Err()
		default:
		}
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return buf.String(), nil
		}
		buf.Write(msg)
		if strings.Contains(buf.String(), "[Terminal exited]") {
			return buf.String(), nil
		}
	}
}

func drainWSHandshake(conn *websocket.Conn, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				return
			}
			return
		}
		if mt == websocket.TextMessage && isControlJSON(msg) {
			continue
		}
		if mt == websocket.BinaryMessage || mt == websocket.TextMessage {
			return
		}
	}
}

func collectWSBinary(conn *websocket.Conn, wait time.Duration) (string, error) {
	return collectWSUntil(conn, wait, nil)
}

func collectWSUntil(conn *websocket.Conn, wait time.Duration, done func(string) bool) (string, error) {
	deadline := time.Now().Add(wait)
	var buf strings.Builder
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			if buf.Len() > 0 && (done == nil || done(buf.String())) {
				return buf.String(), nil
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				if done != nil && done(buf.String()) {
					return buf.String(), nil
				}
				continue
			}
			return buf.String(), nil
		}
		if mt == websocket.TextMessage && isControlJSON(msg) {
			continue
		}
		if mt == websocket.BinaryMessage || mt == websocket.TextMessage {
			appendFilteredWS(&buf, msg)
			if done != nil && done(buf.String()) {
				return buf.String(), nil
			}
		}
	}
	return buf.String(), nil
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