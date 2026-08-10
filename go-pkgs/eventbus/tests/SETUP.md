# Scenario

**Feature**: eventbus L2 harness — Event JSON, HTTP Publisher, optional ListenWS

```
# Event wire contract
caller Event -> json.Marshal/Unmarshal -> field-preserving envelope

# HTTP publish (best-effort)
caller -> Publisher.Publish -> POST {baseURL}/publish JSON (+ optional Bearer)
empty baseURL -> no-op success (no HTTP)

# Optional listen
caller -> ListenWS(wsURL) -> JSON Event frames until ctx cancel
```

## Preconditions

- Package path: `github.com/xhd2015/dot-pkgs/go-pkgs/eventbus` (greenfield; RED until implemented).
- Tests inject `httptest.Server` / gorilla WebSocket upgrade servers — no real hub process.
- Parallel-safe: each leaf owns its mock server; no process env or cwd mutation.
- Locked defaults: publish port `23891`, path `/publish`, ~2s publish timeout.

## Steps

1. Root `Setup` seeds a default fixture `Event` used by most leaves.
2. Branch/leaf `Setup` sets `Op`, publish/token options, and starts HTTP or WS mocks when needed.
3. Root `Run` calls the public package API for the selected `Op`.
4. Leaf `Assert` checks Response fields and/or `req.Capture` HTTP observations.

## Context

- Fixture Event matches the locked wire example shape (`seatalk.message.received`).
- `HTTPCapture` records every request the mock receives (method, path, headers, body).
- Empty base URL leaves must not start an HTTP mock.

```go
import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/eventbus"
)

func fixtureEvent() eventbus.Event {
	return eventbus.Event{
		ID:      "evt_test_001",
		TS:      "2026-08-10T12:34:56.789Z",
		Source:  eventbus.SourceSeatalkLocalBot,
		Type:    eventbus.TypeSeatalkMessageReceived,
		Host:    "test-host",
		Payload: json.RawMessage(`{"text":"hello"}`),
	}
}

// startPublishMock starts an httptest server that records requests and returns statusCode.
// statusCode 0 means 200. Server is closed via t.Cleanup.
func startPublishMock(t *testing.T, statusCode int, capture *HTTPCapture) *httptest.Server {
	t.Helper()
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		capture.add(CapturedRequest{
			Method:        r.Method,
			Path:          r.URL.Path,
			Authorization: r.Header.Get("Authorization"),
			ContentType:   r.Header.Get("Content-Type"),
			Body:          body,
		})
		w.WriteHeader(statusCode)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// startWSEventServer upgrades WebSocket connections and writes each event as a text JSON frame.
func startWSEventServer(t *testing.T, events ...eventbus.Event) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for _, ev := range events {
			b, err := json.Marshal(ev)
			if err != nil {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				return
			}
		}
		// keep connection open briefly so client can read
		time.Sleep(50 * time.Millisecond)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func httpToWSURL(httpURL string) string {
	if strings.HasPrefix(httpURL, "https") {
		return "wss" + strings.TrimPrefix(httpURL, "https")
	}
	if strings.HasPrefix(httpURL, "http") {
		return "ws" + strings.TrimPrefix(httpURL, "http")
	}
	return httpURL
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Event = fixtureEvent()
	return nil
}
```
