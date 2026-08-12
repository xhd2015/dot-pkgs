package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestRunBridgeSendsKeepAlivePings(t *testing.T) {
	var pingCount atomic.Int64
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// Count application-visible pings. gorilla delivers Ping via the
		// handler rather than as ReadMessage frames.
		conn.SetPingHandler(func(appData string) error {
			pingCount.Add(1)
			// Mirror default behavior: reply with Pong so the client read
			// path stays healthy.
			deadline := time.Now().Add(time.Second)
			return conn.WriteControl(websocket.PongMessage, []byte(appData), deadline)
		})
		// Hold the connection open until the client closes (or times out).
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	prev := keepAliveInterval
	keepAliveInterval = 40 * time.Millisecond
	defer func() { keepAliveInterval = prev }()

	// Blocking stdin so the bridge stays up long enough for several pings.
	stdinR, stdinW := io.Pipe()
	defer stdinR.Close()
	defer stdinW.Close()

	done := make(chan error, 1)
	go func() {
		done <- runBridge(conn, stdinR, io.Discard)
	}()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if pingCount.Load() >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got := pingCount.Load(); got < 2 {
		t.Fatalf("expected >= 2 keepalive pings, got %d", got)
	}

	// Unblock stdin so runBridge can exit cleanly via the input path.
	_ = stdinW.Close()
	select {
	case err := <-done:
		if err != nil {
			// Close after peer teardown is acceptable; we only care that
			// pings were sent while the bridge was alive.
			t.Logf("runBridge returned: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runBridge did not return after stdin close")
	}
}

func TestStartKeepAliveDisabledWhenIntervalZero(t *testing.T) {
	// Smoke: must not panic or spawn a ticker that fires.
	stop := make(chan struct{})
	startKeepAlive(&wsWriter{}, 0, stop)
	close(stop)
}
