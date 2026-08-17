package client

import (
	"net/url"
	"testing"
)

func TestTerminalWebSocketURLAttachMode(t *testing.T) {
	t.Parallel()

	got, err := terminalWebSocketURL("http://127.0.0.1:9", ConnectOptions{
		SessionID:      "session-9",
		AttachMode:     "attach",
		AttachSnapshot: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme != "ws" {
		t.Fatalf("scheme: got %q", u.Scheme)
	}
	if u.Path != "/api/terminal" {
		t.Fatalf("path: got %q", u.Path)
	}
	q := u.Query()
	if q.Get("session_id") != "session-9" {
		t.Fatalf("session_id: got %q", q.Get("session_id"))
	}
	if q.Get("attach_mode") != "attach" {
		t.Fatalf("AttachMode must win over AttachSnapshot, got attach_mode=%q", q.Get("attach_mode"))
	}
}

func TestTerminalWebSocketURLAttachSnapshot(t *testing.T) {
	t.Parallel()

	got, err := terminalWebSocketURL("http://127.0.0.1:9", ConnectOptions{
		SessionID:      "session-9",
		AttachSnapshot: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("attach_mode") != "screen" {
		t.Fatalf("AttachSnapshot should set attach_mode=screen, got %q", u.Query().Get("attach_mode"))
	}
}
