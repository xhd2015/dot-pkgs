package portmap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebSocketKeepsRouteInUse(t *testing.T) {
	upgrade := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			kind, data, err := c.ReadMessage()
			if err != nil {
				return
			}
			if c.WriteMessage(kind, data) != nil {
				return
			}
		}
	}))
	defer origin.Close()
	u, _ := url.Parse(origin.URL)
	port, _ := strconv.Atoi(u.Port())
	m, _, advance := testManager(t)
	if _, err := m.Ensure(context.Background(), "p.example.com", port); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	relay := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { defer close(done); m.ServeHTTP(w, r) }))
	defer relay.Close()
	c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(relay.URL, "http")+"/socket", http.Header{"Host": []string{"p.example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if _, data, err := c.ReadMessage(); err != nil || string(data) != "hello" {
		t.Fatalf("%s %v", data, err)
	}
	advance(20 * time.Minute)
	if err := m.Sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if routes := m.List(); len(routes) != 1 || routes[0].Active != 1 {
		t.Fatalf("active websocket expired: %v", routes)
	}
	_ = c.Close()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("proxy did not release upgraded connection")
	}
	advance(10 * time.Minute)
	_ = m.Sweep(context.Background())
	if routes := m.List(); len(routes) != 1 || routes[0].Status != StatusDown {
		t.Fatalf("closed websocket route was removed: %v", routes)
	}
}
