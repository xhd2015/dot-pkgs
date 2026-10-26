package portmap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestRequestCannotReviveExpiredRouteBeforeSweep(t *testing.T) {
	m, p, advance := testManager(t)
	_, err := m.Ensure(context.Background(), "p.example.com", 8421)
	if err != nil {
		t.Fatal(err)
	}
	advance(10 * time.Minute)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, httptest.NewRequest("GET", "https://p.example.com/", nil))
	if rec.Code != 410 || len(m.List()) != 1 {
		t.Fatalf("late request revived or dropped history: %d %v", rec.Code, m.List())
	}
	if _, err := m.Ensure(context.Background(), "p.example.com", 8421); err != nil {
		t.Fatal(err)
	}
	if len(p.published) != 2 || len(m.List()) != 1 {
		t.Fatal("explicit ensure should reactivate expired route")
	}
}
func TestStopCancelsInFlightRequests(t *testing.T) {
	entered := make(chan struct{})
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-r.Context().Done() }))
	defer origin.Close()
	u, _ := url.Parse(origin.URL)
	port, _ := strconv.Atoi(u.Port())
	m, _, _ := testManager(t)
	_, _ = m.Ensure(context.Background(), "p.example.com", port)
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "https://p.example.com/stream", nil))
	}()
	<-entered
	if err := m.Stop(context.Background(), "p.example.com"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("stopped request still forwarding")
	}
	if routes := m.List(); len(routes) != 1 || routes[0].Status != StatusDown {
		t.Fatalf("stopped mapping was removed: %v", routes)
	}
}
