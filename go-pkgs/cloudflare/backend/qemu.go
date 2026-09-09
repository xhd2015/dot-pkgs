package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/qemu"
)

// routesFileName is host-side registry of hostname→origin for the qemu backend
// (so Upsert merges without relying on single-origin shell named-start).
const routesFileName = "backend-routes.json"

type qemuBackend struct {
	m             *qemu.Manager
	log           io.Writer
	defaultTunnel string
}

type routeRegistry struct {
	Tunnel string            `json:"tunnel"`
	Routes map[string]string `json:"routes"` // hostname → origin
}

func newQemuBackend(s Select) *qemuBackend {
	m := s.QemuManager
	if m == nil {
		m = &qemu.Manager{
			Cfg:  qemu.WorkDevConfig(),
			Host: qemu.LocalHost{},
		}
	}
	tunnel := s.DefaultTunnel
	if tunnel == "" {
		tunnel = DefaultQemuTunnel
	}
	logw := s.Log
	if logw == nil {
		logw = io.Discard
	}
	return &qemuBackend{m: m, log: logw, defaultTunnel: tunnel}
}

func (q *qemuBackend) Kind() Kind { return KindQemu }

func (q *qemuBackend) OriginForHostPort(port int) string {
	return qemu.GuestOriginForPort(port)
}

func (q *qemuBackend) Status(ctx context.Context) (Status, error) {
	_ = ctx
	st := Status{Kind: KindQemu}
	kv, err := q.m.CFStatus("")
	if err != nil && kv["qemu_alive"] == "" {
		st.Detail = err.Error()
		return st, nil
	}
	if kv["qemu_alive"] != "yes" {
		st.Detail = "qemu guest not running"
		return st, nil
	}
	st.Installed = kv["cf_bin"] == "ok"
	st.Authenticated = kv["cert"] == "ok"
	if !st.Installed {
		st.Detail = "guest cloudflared binary missing"
		return st, nil
	}
	if !st.Authenticated {
		st.Detail = "guest cert.pem missing (qemu cloudflared login)"
		return st, nil
	}
	st.Available = true
	st.Detail = "qemu guest cloudflared ready"
	return st, nil
}

func (q *qemuBackend) EnsureReady(ctx context.Context) error {
	if _, err := q.m.EnsureRunning(); err != nil {
		return fmt.Errorf("ensure qemu: %w", err)
	}
	st, err := q.Status(ctx)
	if err != nil {
		return err
	}
	if !st.Available {
		// Try ensuring cf binary via status/start path (CFStatus ensure is in script start/named).
		kv, _ := q.m.CFStatus("")
		if kv["cert"] != "ok" {
			return fmt.Errorf("qemu cloudflared not ready: %s", st.Detail)
		}
		if kv["cf_bin"] != "ok" {
			return fmt.Errorf("qemu cloudflared not ready: %s", st.Detail)
		}
		return fmt.Errorf("qemu cloudflared not ready: %s", st.Detail)
	}
	return nil
}

func (q *qemuBackend) UpsertRoute(ctx context.Context, r Route) (string, error) {
	if err := validateRoute(r); err != nil {
		return "", err
	}
	if err := q.EnsureReady(ctx); err != nil {
		return "", err
	}
	tunnel := r.Tunnel
	if tunnel == "" {
		tunnel = q.defaultTunnel
	}
	reg, err := q.loadRegistry(tunnel)
	if err != nil {
		return "", err
	}
	reg.Tunnel = tunnel
	if reg.Routes == nil {
		reg.Routes = map[string]string{}
	}
	q.mergeGuestIngress(reg)
	prev := reg.Routes[r.Hostname]
	reg.Routes[r.Hostname] = r.Origin
	if err := q.saveRegistry(reg); err != nil {
		return "", err
	}
	// Skip guest CF restart when the mapping is already live — bouncing the
	// connector drops the public tunnel (and remote-agent HTTP) mid-request.
	if prev == r.Origin {
		fmt.Fprintf(q.log, "qemu backend: route %s already -> %s; skip apply\n", r.Hostname, r.Origin)
		return publicURL(r.Hostname), nil
	}
	if err := q.applyRegistry(reg); err != nil {
		return "", err
	}
	return publicURL(r.Hostname), nil
}

func (q *qemuBackend) RemoveRoute(ctx context.Context, tunnel, hostname string) error {
	_ = ctx
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if tunnel == "" {
		tunnel = q.defaultTunnel
	}
	reg, err := q.loadRegistry(tunnel)
	if err != nil {
		return err
	}
	if reg.Routes == nil {
		return nil
	}
	delete(reg.Routes, hostname)
	reg.Tunnel = tunnel
	if err := q.saveRegistry(reg); err != nil {
		return err
	}
	if len(reg.Routes) == 0 {
		fmt.Fprintf(q.log, "qemu backend: no routes left; stopping guest cloudflared\n")
		return q.m.CFStop()
	}
	return q.applyRegistry(reg)
}

func (q *qemuBackend) routesPath() string {
	cfg := q.m.Cfg
	// normalized defaults applied inside Guest* scripts; Dir may be set on manager
	dir := cfg.Dir
	if dir == "" {
		dir = qemu.WorkDevConfig().Dir
	}
	return filepath.Join(dir, routesFileName)
}

func (q *qemuBackend) loadRegistry(tunnel string) (*routeRegistry, error) {
	path := q.routesPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &routeRegistry{Tunnel: tunnel, Routes: map[string]string{}}, nil
		}
		return nil, err
	}
	var reg routeRegistry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if reg.Routes == nil {
		reg.Routes = map[string]string{}
	}
	if reg.Tunnel == "" {
		reg.Tunnel = tunnel
	}
	return &reg, nil
}

func (q *qemuBackend) saveRegistry(reg *routeRegistry) error {
	path := q.routesPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func (q *qemuBackend) applyRegistry(reg *routeRegistry) error {
	if reg.Tunnel == "" {
		return fmt.Errorf("tunnel name is required")
	}
	hosts := make([]string, 0, len(reg.Routes))
	for h := range reg.Routes {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)
	fmt.Fprintf(q.log, "qemu backend: applying %d routes on tunnel %s\n", len(hosts), reg.Tunnel)
	return q.m.CFNamedApply(reg.Tunnel, hosts, reg.Routes, "")
}

// mergeGuestIngress pulls hostname→origin from the live guest yml so Upsert
// does not drop sibling routes that were configured outside this registry.
func (q *qemuBackend) mergeGuestIngress(reg *routeRegistry) {
	if reg.Tunnel == "" {
		return
	}
	path := "/root/.cloudflared/" + reg.Tunnel + ".yml"
	body, err := q.m.GuestReadFile(path)
	if err != nil || body == "" {
		return
	}
	for h, o := range parseIngressHostOrigins(body) {
		if _, exists := reg.Routes[h]; !exists {
			reg.Routes[h] = o
		}
	}
}

func parseIngressHostOrigins(yml string) map[string]string {
	out := map[string]string{}
	var curHost string
	for _, line := range strings.Split(yml, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "- hostname:") {
			curHost = strings.TrimSpace(strings.TrimPrefix(trim, "- hostname:"))
			continue
		}
		if curHost != "" && strings.HasPrefix(trim, "service:") {
			svc := strings.TrimSpace(strings.TrimPrefix(trim, "service:"))
			if svc != "" && !strings.HasPrefix(svc, "http_status:") {
				out[curHost] = svc
			}
			curHost = ""
		}
	}
	return out
}
