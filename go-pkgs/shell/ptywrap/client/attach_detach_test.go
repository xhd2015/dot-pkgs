package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestAttachCtrlBracketSendsDetachKeep(t *testing.T) {
	got := make(chan string, 4)
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"session_id","session_id":"session-1"}`))
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var msg struct {
				Type string `json:"type"`
			}
			if json.Unmarshal(data, &msg) == nil && msg.Type != "" {
				got <- msg.Type
			}
		}
	}))
	t.Cleanup(srv.Close)

	stdinR, stdinW := io.Pipe()
	t.Cleanup(func() { _ = stdinR.Close(); _ = stdinW.Close() })

	done := make(chan AttachResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := AttachWithIO(&Client{BaseURL: srv.URL}, ConnectOptions{
			SessionID:    "session-1",
			Wait:         true,
			SkipTTYCheck: true,
		}, stdinR, io.Discard, io.Discard)
		done <- res
		errCh <- err
	}()

	time.Sleep(50 * time.Millisecond)
	if _, err := stdinW.Write([]byte{0x1d}); err != nil {
		t.Fatal(err)
	}

	select {
	case typ := <-got:
		if typ != "detach_keep" {
			t.Fatalf("control type: got %q, want detach_keep", typ)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for detach_keep")
	}

	select {
	case res := <-done:
		if err := <-errCh; err != nil {
			t.Fatalf("AttachWithIO: %v", err)
		}
		if !res.Detached {
			t.Fatal("expected Detached")
		}
		if res.SessionID != "session-1" {
			t.Fatalf("SessionID: got %q", res.SessionID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("AttachWithIO did not return after Ctrl-]")
	}
}
