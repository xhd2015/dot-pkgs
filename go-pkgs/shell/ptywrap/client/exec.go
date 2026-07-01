package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

// ExecOptions configures a one-shot exec WebSocket session against the
// /api/exec/ws endpoint. Unlike Attach, an exec session runs a single
// subprocess and returns its exit code rather than a persistent session.
type ExecOptions struct {
	// Argv is the command and its arguments; Argv[0] is the binary.
	Argv []string

	// Stdin is read and forwarded to the remote process. Defaults to
	// os.Stdin when nil.
	Stdin io.Reader

	// Stdout receives the remote process's stdout. Defaults to os.Stdout
	// when nil.
	Stdout io.Writer

	// AuthToken, when set, overrides c.AuthToken for this call.
	AuthToken string

	// SkipTTYCheck disables raw-mode terminal handling. When false (the
	// default), RunExec enters raw mode and forwards SIGWINCH resize
	// events if both stdin and stdout are TTYs. Tests set this to true
	// to drive the client with in-memory pipes.
	SkipTTYCheck bool
}

type execRequest struct {
	Argv []string `json:"argv"`
}

type execServerMessage struct {
	Type    string `json:"type"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type execReadResult struct {
	exitCode int
	err      error
}

// RunExec dials /api/exec/ws on the daemon, runs argv as a subprocess on
// the remote side, bridges local IO to it, and returns the remote exit
// code. A non-zero exit code is returned as a normal int (not a Go error);
// a Go error is returned only when the connection or stream fails.
func RunExec(c *Client, opts ExecOptions) (int, error) {
	stdin := opts.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	interactive := false
	if !opts.SkipTTYCheck {
		interactive = isInteractiveIO(stdin, stdout)
	}

	wsURL, err := execWebSocketURL(c.BaseURL)
	if err != nil {
		return 0, err
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
		return 0, terminalDialError(err, resp)
	}
	defer conn.Close()

	writer := &wsWriter{conn: conn}
	if err := writer.writeJSON(execRequest{Argv: opts.Argv}); err != nil {
		return 0, fmt.Errorf("send exec request: %w", err)
	}

	if interactive {
		if stdinFile, ok := stdin.(*os.File); ok {
			if state, err := term.MakeRaw(int(stdinFile.Fd())); err == nil {
				defer term.Restore(int(stdinFile.Fd()), state)
			}
		}
		if stdoutFile, ok := stdout.(*os.File); ok {
			_ = sendTerminalResize(writer, stdoutFile)
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

	readerCh := make(chan execReadResult, 1)
	go func() {
		code, err := readExecOutput(conn, stdout)
		readerCh <- execReadResult{exitCode: code, err: err}
	}()

	stdinErrCh := make(chan error, 1)
	go func() {
		stdinErrCh <- forwardTerminalInput(writer, stdin)
	}()

	for {
		select {
		case res := <-readerCh:
			if res.err != nil {
				return 0, res.err
			}
			return res.exitCode, nil
		case err := <-stdinErrCh:
			if err == nil || err == io.EOF {
				continue
			}
			_ = conn.Close()
			res := <-readerCh
			if res.err != nil {
				return 0, res.err
			}
			return 0, err
		}
	}
}

// readExecOutput pumps server messages to stdout until an "exit" or
// "error" control message terminates the session.
func readExecOutput(conn *websocket.Conn, stdout io.Writer) (int, error) {
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			return 0, normalizeTerminalReadError(err)
		}
		switch msgType {
		case websocket.BinaryMessage:
			if _, err := stdout.Write(data); err != nil {
				return 0, err
			}
		case websocket.TextMessage:
			var msg execServerMessage
			if err := json.Unmarshal(data, &msg); err == nil && msg.Type != "" {
				switch msg.Type {
				case "exit":
					return msg.Code, nil
				case "error":
					if msg.Message == "" {
						return 0, fmt.Errorf("remote exec error")
					}
					return 0, fmt.Errorf("%s", msg.Message)
				default:
					if msg.Message != "" {
						if _, err := stdout.Write([]byte(msg.Message)); err != nil {
							return 0, err
						}
					}
					continue
				}
			}
			if _, err := stdout.Write(data); err != nil {
				return 0, err
			}
		}
	}
}

func execWebSocketURL(base string) (string, error) {
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
	u.Path = "/api/exec/ws"
	u.RawQuery = ""
	return u.String(), nil
}
