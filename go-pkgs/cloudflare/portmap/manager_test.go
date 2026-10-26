package portmap

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakePublisher struct {
	mu                 sync.Mutex
	published, removed []string
	failRemove         bool
}

func (p *fakePublisher) Publish(_ context.Context, h, o string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.published = append(p.published, h)
	return nil
}
func (p *fakePublisher) Unpublish(_ context.Context, h string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.failRemove {
		return fmt.Errorf("cleanup failed")
	}
	p.removed = append(p.removed, h)
	return nil
}
func testManager(t *testing.T) (*Manager, *fakePublisher, func(time.Duration)) {
	t.Helper()
	var stamp atomic.Int64
	stamp.Store(time.Now().UnixNano())
	p := &fakePublisher{}
	m, err := New(Options{Publisher: p, Now: func() time.Time { return time.Unix(0, stamp.Load()) }})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m, p, func(d time.Duration) { stamp.Add(int64(d)) }
}
func TestIdleExpiresWithoutManagementRenewal(t *testing.T) {
	m, p, advance := testManager(t)
	ctx := context.Background()
	_, err := m.Ensure(ctx, "p-8421.example.com", 8421)
	if err != nil {
		t.Fatal(err)
	}
	advance(9 * time.Minute)
	_ = m.List()
	_, err = m.Ensure(ctx, "p-8421.example.com", 8421)
	if err != nil {
		t.Fatal(err)
	}
	advance(time.Minute)
	if err = m.Sweep(ctx); err != nil {
		t.Fatal(err)
	}
	routes := m.List()
	if len(routes) != 1 || routes[0].Status != StatusDown || len(p.published) != 1 || len(p.removed) != 1 {
		t.Fatalf("routes=%v published=%v removed=%v", routes, p.published, p.removed)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "https://p-8421.example.com/", nil)
	m.ServeHTTP(rec, req)
	if rec.Code != 410 {
		t.Fatal(rec.Code)
	}
}
func TestTrafficResetsIdleAndPreservesPath(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, r.URL.String()) }))
	defer origin.Close()
	u, _ := url.Parse(origin.URL)
	n, _ := strconv.Atoi(u.Port())
	m, _, advance := testManager(t)
	_, err := m.Ensure(context.Background(), "p.example.com", n)
	if err != nil {
		t.Fatal(err)
	}
	advance(9 * time.Minute)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, httptest.NewRequest("GET", "https://p.example.com/app?x=1", nil))
	if rec.Code != 200 || rec.Body.String() != "/app?x=1" {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	advance(2 * time.Minute)
	_ = m.Sweep(context.Background())
	if len(m.List()) != 1 {
		t.Fatal("request did not renew idle")
	}
	advance(8 * time.Minute)
	_ = m.Sweep(context.Background())
	if routes := m.List(); len(routes) != 1 || routes[0].Status != StatusDown {
		t.Fatalf("route was removed instead of marked down: %v", routes)
	}
}
func TestActiveStreamDoesNotExpire(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-release; fmt.Fprint(w, "done") }))
	defer origin.Close()
	u, _ := url.Parse(origin.URL)
	n, _ := strconv.Atoi(u.Port())
	m, _, advance := testManager(t)
	_, _ = m.Ensure(context.Background(), "p.example.com", n)
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "https://p.example.com/stream", nil))
	}()
	<-entered
	advance(20 * time.Minute)
	_ = m.Sweep(context.Background())
	routes := m.List()
	if len(routes) != 1 || routes[0].Active != 1 || routes[0].ExpiresAt != nil {
		t.Errorf("active stream expired: %v", routes)
	}
	close(release)
	<-done
	advance(9 * time.Minute)
	_ = m.Sweep(context.Background())
	if len(m.List()) != 1 {
		t.Fatal("idle grace should start when stream ends")
	}
	advance(time.Minute)
	_ = m.Sweep(context.Background())
	if routes := m.List(); len(routes) != 1 || routes[0].Status != StatusDown {
		t.Fatalf("ended stream was removed: %v", routes)
	}
}
func TestConcurrentEnsureAndSiblingStop(t *testing.T) {
	m, p, _ := testManager(t)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := m.Ensure(context.Background(), "p.example.com", 8421); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if len(p.published) != 1 {
		t.Fatalf("duplicate publication %v", p.published)
	}
	_, _ = m.Ensure(context.Background(), "other.example.com", 8422)
	_ = m.Stop(context.Background(), "p.example.com")
	var up, down int
	for _, route := range m.List() {
		if route.Port == 8422 && route.Status == StatusUp {
			up++
		}
		if route.Port == 8421 && route.Status == StatusDown {
			down++
		}
	}
	if up != 1 || down != 1 {
		t.Fatal(m.List())
	}
	if err := m.Forget("p.example.com"); err != nil || len(m.List()) != 1 {
		t.Fatalf("forget down history: %v %v", m.List(), err)
	}
	if err := m.Forget("other.example.com"); err == nil {
		t.Fatal("removed an up mapping from history")
	}
}
func TestFailedCleanupStaysVisibleAndRetries(t *testing.T) {
	m, p, advance := testManager(t)
	_, _ = m.Ensure(context.Background(), "p.example.com", 8421)
	p.failRemove = true
	advance(10 * time.Minute)
	if m.Sweep(context.Background()) == nil {
		t.Fatal("want cleanup error")
	}
	if len(m.List()) != 1 || m.List()[0].Error == "" {
		t.Fatal(m.List())
	}
	if m.List()[0].Status != StatusUp {
		t.Fatal(m.List())
	}
	p.failRemove = false
	if err := m.Sweep(context.Background()); err != nil || len(m.List()) != 1 || m.List()[0].Status != StatusDown {
		t.Fatalf("%v %v", m.List(), err)
	}
}
