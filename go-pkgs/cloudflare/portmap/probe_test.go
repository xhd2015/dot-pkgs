package portmap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestRelayProofDoesNotTouchAppOrIdle(t *testing.T) {
	var requests atomic.Int64
	app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Header.Get(ProbeSecretHeader) != "" {
			t.Error("probe credential reached app")
		}
		http.Error(w, "application 404", http.StatusNotFound)
	}))
	defer app.Close()
	u, _ := url.Parse(app.URL)
	port, _ := strconv.Atoi(u.Port())
	var clock atomic.Int64
	clock.Store(time.Now().UnixNano())
	m, err := New(Options{Publisher: &fakePublisher{}, ProbeSecret: "secret", Now: func() time.Time { return time.Unix(0, clock.Load()) }})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.Ensure(context.Background(), "app.example.com", port); err != nil {
		t.Fatal(err)
	}
	clock.Add(int64(9 * time.Minute))
	probe := httptest.NewRequest("GET", "https://app.example.com/", nil)
	probe.Header.Set(ProbeSecretHeader, "secret")
	probe.Header.Set(ProbeNonceHeader, "fresh")
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, probe)
	if rec.Code != http.StatusNoContent || rec.Header().Get(ProbeReplyHeader) != "fresh" || requests.Load() != 0 {
		t.Fatalf("probe: %d %v requests=%d", rec.Code, rec.Header(), requests.Load())
	}
	clock.Add(int64(time.Minute))
	if err := m.Sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if m.List()[0].Status != StatusDown {
		t.Fatal("probe extended idle")
	}
}

func TestApplication404IsNotAProbeAndProbeHeadersAreStripped(t *testing.T) {
	app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(ProbeSecretHeader) != "" {
			t.Error("secret header leaked")
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer app.Close()
	u, _ := url.Parse(app.URL)
	port, _ := strconv.Atoi(u.Port())
	m, err := New(Options{Publisher: &fakePublisher{}, ProbeSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.Ensure(context.Background(), "app.example.com", port); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "https://app.example.com/nope", nil)
	r.Header.Set(ProbeSecretHeader, "wrong")
	r.Header.Set(ProbeNonceHeader, "nonce")
	w := httptest.NewRecorder()
	m.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound || w.Header().Get(ProbeReplyHeader) != "" {
		t.Fatalf("status=%d headers=%v", w.Code, w.Header())
	}
}
