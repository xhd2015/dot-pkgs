// Package portmap publishes local HTTP services through a traffic-metered relay.
// Idle is measured from real application requests, never management polling.
package portmap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// Publisher exposes the relay at hostname; implementations must not overwrite
// unrelated DNS records. Unpublish releases only this manager's hostname.
type Publisher interface {
	Publish(context.Context, string, string) error
	Unpublish(context.Context, string) error
}

type Options struct {
	Publisher   Publisher
	IdleTimeout time.Duration
	Now         func() time.Time
	// UseUnixSocket prevents a surviving connector from targeting a reused TCP
	// port after this process crashes. ManagedPublisher supports unix origins.
	UseUnixSocket bool
	// OnChange receives the history after a route is published, stopped, or removed.
	// It must not publish anything.
	OnChange func([]Route)
}

const (
	StatusUp   = "up"
	StatusDown = "down"
)

type Route struct {
	Hostname      string     `json:"hostname"`
	Port          int        `json:"port"`
	PublicURL     string     `json:"publicUrl"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastRequestAt time.Time  `json:"lastRequestAt"`
	Active        int        `json:"active"`
	Status        string     `json:"status"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	Error         string     `json:"error,omitempty"`
}

type entry struct {
	Route
	proxy    *httputil.ReverseProxy
	lastUsed time.Time
	live     bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// Manager serializes publication/removal, while traffic takes only the brief
// state mutex. A request accepted before expiry keeps its stream alive.
type Manager struct {
	op        sync.Mutex
	mu        sync.Mutex
	routes    map[string]*entry
	publisher Publisher
	idle      time.Duration
	now       func() time.Time
	listener  net.Listener
	server    *http.Server
	origin    string
	closed    bool
	cancel    context.CancelFunc
	cleanup   func()
	onChange  func([]Route)
}

func New(opts Options) (*Manager, error) {
	if opts.Publisher == nil {
		return nil, fmt.Errorf("publisher is required")
	}
	if opts.IdleTimeout <= 0 {
		opts.IdleTimeout = 10 * time.Minute
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	l, origin, cleanup, err := listenRelay(opts.UseUnixSocket)
	if err != nil {
		return nil, err
	}
	m := &Manager{routes: map[string]*entry{}, publisher: opts.Publisher, idle: opts.IdleTimeout, now: opts.Now, listener: l, origin: origin, cleanup: cleanup, onChange: opts.OnChange}
	m.server = &http.Server{Handler: m, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = m.server.Serve(l) }()
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				call, stop := context.WithTimeout(ctx, 30*time.Second)
				_ = m.Sweep(call)
				stop()
			}
		}
	}()
	return m, nil
}

func (m *Manager) Ensure(ctx context.Context, hostname string, port int) (Route, error) {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	if hostname == "" || strings.ContainsAny(hostname, "/: \\*\r\n") {
		return Route{}, fmt.Errorf("invalid hostname")
	}
	if port < 1 || port > 65535 {
		return Route{}, fmt.Errorf("port must be 1–65535")
	}
	m.op.Lock()
	defer m.op.Unlock()
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return Route{}, fmt.Errorf("port mapper is closed")
	}
	var cancel context.CancelFunc
	if e := m.routes[hostname]; e != nil {
		if e.Port != port {
			m.mu.Unlock()
			return Route{}, fmt.Errorf("hostname already maps to another port")
		}
		if e.live && (e.Active > 0 || m.now().Sub(e.lastUsed) < m.idle) {
			out := m.snapshot(e)
			m.mu.Unlock()
			return out, nil
		}
		e.live = false
		cancel = e.cancel
		e.cancel = nil
	}
	if cancel != nil {
		m.mu.Unlock()
		cancel()
		m.mu.Lock()
		if m.closed {
			m.mu.Unlock()
			return Route{}, fmt.Errorf("port mapper is closed")
		}
	}
	now := m.now()
	created := now
	if old := m.routes[hostname]; old != nil && !old.CreatedAt.IsZero() {
		created = old.CreatedAt
	}
	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	e := &entry{Route: Route{Hostname: hostname, Port: port, PublicURL: "https://" + hostname, CreatedAt: created, LastRequestAt: now, Status: StatusUp}, lastUsed: now}
	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.proxy = httputil.NewSingleHostReverseProxy(target)
	director := e.proxy.Director
	e.proxy.Director = func(request *http.Request) {
		publicHost := request.Host
		director(request)
		// Local dev servers commonly restrict Host to loopback. Preserve the
		// public identity in forwarded headers instead of disabling those checks.
		request.Host = target.Host
		request.Header.Set("X-Forwarded-Host", publicHost)
		request.Header.Set("X-Forwarded-Proto", "https")
	}
	// Do not cache app errors or HTML at the Cloudflare edge. Upgraded streams
	// retain ServeHTTP ownership until their connection is closed.
	e.proxy.ModifyResponse = func(r *http.Response) error { r.Header.Set("Cache-Control", "no-store"); return nil }
	m.routes[hostname] = e
	m.mu.Unlock()
	if err := m.publisher.Publish(ctx, hostname, m.origin); err != nil {
		m.mu.Lock()
		e.Error = err.Error()
		e.Status = StatusUp
		m.mu.Unlock()
		m.changed()
		return Route{}, err
	}
	m.mu.Lock()
	e.live = true
	e.Status = StatusUp
	e.Error = ""
	out := m.snapshot(e)
	m.mu.Unlock()
	m.changed()
	return out, nil
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := strings.ToLower(r.Host)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	m.mu.Lock()
	e := m.routes[host]
	// Enforce the deadline on admission too: a delayed sweeper or a slow
	// unrelated publication must never let expired traffic revive a route.
	if e != nil && e.live && e.Active == 0 && m.now().Sub(e.lastUsed) >= m.idle {
		e.live = false
		e.cancel()
	}
	if e == nil || !e.live || m.closed {
		m.mu.Unlock()
		http.Error(w, "port mapping is not active; reopen it through the managing application", http.StatusGone)
		return
	}
	e.Active++
	e.LastRequestAt = m.now()
	e.lastUsed = e.LastRequestAt
	m.mu.Unlock()
	requestCtx, cancel := context.WithCancel(r.Context())
	stop := context.AfterFunc(e.ctx, cancel)
	defer func() { stop(); cancel() }()
	defer func() { m.mu.Lock(); e.Active--; e.lastUsed = m.now(); m.mu.Unlock() }()
	e.proxy.ServeHTTP(w, r.WithContext(requestCtx))
}

func (m *Manager) snapshot(e *entry) Route {
	r := e.Route
	if r.Status == "" {
		if e.live {
			r.Status = StatusUp
		} else {
			r.Status = StatusDown
		}
	}
	if r.Status != StatusUp {
		r.Active = 0
		r.ExpiresAt = nil
		return r
	}
	if r.Active == 0 {
		deadline := e.lastUsed.Add(m.idle)
		r.ExpiresAt = &deadline
	}
	return r
}
func (m *Manager) List() []Route {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Route, 0, len(m.routes))
	for _, e := range m.routes {
		out = append(out, m.snapshot(e))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Hostname < out[j].Hostname })
	return out
}

// Sweep blocks new traffic atomically before unpublishing an idle route. Failed
// cleanup remains visible and is retried; Ensure cannot race removal.
func (m *Manager) Sweep(ctx context.Context) error {
	m.op.Lock()
	defer m.op.Unlock()
	m.mu.Lock()
	var hosts []string
	for host, e := range m.routes {
		if e.Status == StatusDown {
			continue
		}
		if e.Active == 0 && m.now().Sub(e.lastUsed) >= m.idle {
			e.live = false
			hosts = append(hosts, host)
		}
	}
	m.mu.Unlock()
	var first error
	for _, host := range hosts {
		if err := m.retire(ctx, host); err != nil && first == nil {
			first = err
		}
	}
	if len(hosts) > 0 {
		m.changed()
	}
	return first
}
func (m *Manager) Stop(ctx context.Context, host string) error {
	m.op.Lock()
	defer m.op.Unlock()
	m.mu.Lock()
	if e := m.routes[host]; e != nil {
		e.live = false
	}
	m.mu.Unlock()
	err := m.retire(ctx, host)
	m.changed()
	return err
}

// Forget removes a Down history row. An Up mapping must be stopped first.
func (m *Manager) Forget(host string) error {
	m.op.Lock()
	defer m.op.Unlock()
	m.mu.Lock()
	e := m.routes[host]
	if e == nil {
		m.mu.Unlock()
		return nil
	}
	if e.live || e.Status == StatusUp {
		m.mu.Unlock()
		return fmt.Errorf("stop the mapping before removing its history")
	}
	delete(m.routes, host)
	m.mu.Unlock()
	m.changed()
	return nil
}

// Remember records history without publishing. Every row is Down.
func (m *Manager) Remember(routes []Route) {
	m.op.Lock()
	defer m.op.Unlock()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, route := range routes {
		host := strings.ToLower(strings.TrimSpace(route.Hostname))
		if host == "" || route.Port < 1 || route.Port > 65535 {
			continue
		}
		if e := m.routes[host]; e != nil && (e.live || e.Status == StatusUp) {
			continue
		}
		route.Hostname = host
		route.Status = StatusDown
		route.Active = 0
		route.ExpiresAt = nil
		route.Error = ""
		if route.PublicURL == "" {
			route.PublicURL = "https://" + host
		}
		m.routes[host] = &entry{Route: route, lastUsed: route.LastRequestAt}
	}
}

// retire unpublishes a route and keeps it as Down. A failed unpublish stays Up.
func (m *Manager) retire(ctx context.Context, host string) error {
	m.mu.Lock()
	var cancel context.CancelFunc
	if e := m.routes[host]; e != nil {
		cancel = e.cancel
		e.cancel = nil
	}
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if err := m.publisher.Unpublish(ctx, host); err != nil {
		m.mu.Lock()
		if e := m.routes[host]; e != nil {
			e.Error = err.Error()
			e.Status = StatusUp
		}
		m.mu.Unlock()
		return err
	}
	m.mu.Lock()
	if e := m.routes[host]; e != nil {
		e.live = false
		e.Status = StatusDown
		e.Active = 0
		e.Error = ""
		e.ExpiresAt = nil
	}
	m.mu.Unlock()
	return nil
}
func (m *Manager) changed() {
	if m.onChange == nil {
		return
	}
	m.onChange(m.List())
}
func (m *Manager) Close() error {
	m.cancel()
	m.op.Lock()
	defer m.op.Unlock()
	m.mu.Lock()
	m.closed = true
	hosts := make([]string, 0, len(m.routes))
	for host, e := range m.routes {
		e.live = false
		if e.cancel != nil {
			e.cancel()
		}
		hosts = append(hosts, host)
	}
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var first error
	for _, host := range hosts {
		if err := m.retire(ctx, host); err != nil && first == nil {
			first = err
		}
	}
	m.changed()
	if err := m.server.Close(); err != nil && first == nil {
		first = err
	}
	m.cleanup()
	return first
}
