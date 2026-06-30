package clienttest

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
)

// Request is the doctest harness request for ptywrap client tests.
type Request struct {
	Phase string

	ServerBase string
	Target     string
	Sessions   []ptywrap.SessionInfo

	UsePipeStdin  bool
	UsePipeStdout bool

	ConnectOpts ptyclient.ConnectOptions
}

// Response is the doctest harness response for ptywrap client tests.
type Response struct {
	Resolved   *ptywrap.SessionInfo
	AttachID   string
	AttachErr  string
	ResolveErr string
}

// Run executes a ptywrap client doctest phase.
func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Phase {
	case "attach-requires-tty":
		return runAttachRequiresTTY(t, req)
	case "attach-captures-id":
		return runAttachCapturesID(t, req)
	case "resolve-id", "resolve-name", "resolve-ambiguous":
		return runResolveTarget(t, req)
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}

func runAttachRequiresTTY(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	base := req.ServerBase
	if base == "" {
		base, _ = StartClientTestServer(t, req.Sessions)
	}
	c := ptyclient.NewClient(base)
	opts := req.ConnectOpts
	if opts.SessionID == "" {
		opts.SessionID = "session-1"
	}
	opts.SkipTTYCheck = false

	stdin, stdout, stderr := os.Stdin, os.Stdout, os.Stderr
	if req.UsePipeStdin {
		r, w, err := os.Pipe()
		if err != nil {
			return nil, err
		}
		defer r.Close()
		defer w.Close()
		stdin = r
	}
	if req.UsePipeStdout {
		r, w, err := os.Pipe()
		if err != nil {
			return nil, err
		}
		defer r.Close()
		defer w.Close()
		stdout = w
		stderr = w
	}

	_, err := ptyclient.AttachWithIO(c, opts, stdin, stdout, stderr)
	if err != nil {
		resp.AttachErr = err.Error()
	}
	return resp, nil
}

func runAttachCapturesID(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	base := req.ServerBase
	if base == "" {
		base, _ = StartClientTestServer(t, req.Sessions)
	}
	c := ptyclient.NewClient(base)
	ttyR, ttyW, err := OpenFakeTTY(t)
	if err != nil {
		return nil, err
	}
	defer ttyR.Close()
	defer ttyW.Close()

	opts := req.ConnectOpts
	opts.SkipTTYCheck = true
	opts.Wait = false
	if opts.SessionID == "" {
		opts.SessionID = "session-42"
	}
	result, err := ptyclient.AttachWithIO(c, opts, ttyR, ttyW, ttyW)
	if err != nil {
		resp.AttachErr = err.Error()
		return resp, nil
	}
	resp.AttachID = result.SessionID
	return resp, nil
}

func runResolveTarget(t *testing.T, req *Request) (*Response, error) {
	resp := &Response{}
	c := ptyclient.NewClient(req.ServerBase)
	c.SetTestSessions(req.Sessions)
	s, err := ptyclient.ResolveTarget(c, req.Target)
	if err != nil {
		resp.ResolveErr = err.Error()
		return resp, nil
	}
	resp.Resolved = s
	return resp, nil
}

// StartClientTestServer starts a test HTTP server with optional seeded sessions.
func StartClientTestServer(t *testing.T, sessions []ptywrap.SessionInfo) (string, func()) {
	t.Helper()
	mux := http.NewServeMux()
	mgr := ptywrap.NewManager()
	if len(sessions) > 0 {
		if err := ptywrap.RegisterTestSessions(mgr, sessions); err != nil {
			t.Fatal(err)
		}
	}
	ptywrap.RegisterAPIWithManager(mux, mgr)
	srv := httptest.NewServer(mux)
	return srv.URL, srv.Close
}

// OpenFakeTTY returns a PTY usable as stdin/stdout in tests.
func OpenFakeTTY(t *testing.T) (stdin *os.File, stdout *os.File, err error) {
	t.Helper()
	ptmx, err := ptyclient.OpenTestPTY()
	if err != nil {
		return nil, nil, err
	}
	return ptmx, ptmx, nil
}

// WSSendSessionID writes a session_id JSON message to a websocket.
func WSSendSessionID(conn *websocket.Conn, id string) error {
	msg := []byte(`{"type":"session_id","session_id":"` + id + `"}`)
	return conn.WriteMessage(websocket.TextMessage, msg)
}

// WaitForTCP waits until a TCP address accepts connections.
func WaitForTCP(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", addr)
}