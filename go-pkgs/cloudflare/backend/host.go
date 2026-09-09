package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
)

type hostBackend struct {
	configDir     string
	runner        cloudflare.CommandRunner
	log           io.Writer
	defaultTunnel string
}

func newHostBackend(s Select) *hostBackend {
	tunnel := s.DefaultTunnel
	if tunnel == "" {
		tunnel = cloudflare.DefaultTunnelName
	}
	logw := s.Log
	if logw == nil {
		logw = io.Discard
	}
	return &hostBackend{
		configDir:     s.HostConfigDir,
		runner:        s.HostRunner,
		log:           logw,
		defaultTunnel: tunnel,
	}
}

func (h *hostBackend) Kind() Kind { return KindHost }

func (h *hostBackend) OriginForHostPort(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d", port)
}

func (h *hostBackend) Status(ctx context.Context) (Status, error) {
	_ = ctx
	st := Status{Kind: KindHost}
	info, err := cloudflare.Status(h.runner)
	if err != nil {
		st.Detail = err.Error()
		return st, nil
	}
	st.Installed = info.Installed
	st.Detail = info.Detail
	if !info.Installed {
		return st, nil
	}
	// Auth: cert.pem under config dir (or default ~/.cloudflared).
	cfgDir := h.configDir
	if cfgDir == "" {
		var derr error
		cfgDir, derr = cloudflare.DefaultConfigDir()
		if derr != nil {
			st.Detail = derr.Error()
			return st, nil
		}
	}
	cert := filepath.Join(cfgDir, "cert.pem")
	if fi, err := os.Stat(cert); err == nil && fi.Size() > 0 {
		st.Authenticated = true
		st.Available = true
		st.Detail = "host cloudflared ready"
	} else {
		st.Detail = "cloudflared installed but cert.pem missing (run cloudflared tunnel login)"
	}
	return st, nil
}

func (h *hostBackend) EnsureReady(ctx context.Context) error {
	st, err := h.Status(ctx)
	if err != nil {
		return err
	}
	if !st.Available {
		return fmt.Errorf("host cloudflared not ready: %s", st.Detail)
	}
	return nil
}

func (h *hostBackend) UpsertRoute(ctx context.Context, r Route) (string, error) {
	_ = ctx
	if err := validateRoute(r); err != nil {
		return "", err
	}
	if err := h.EnsureReady(ctx); err != nil {
		return "", err
	}
	tunnel := r.Tunnel
	if tunnel == "" {
		tunnel = h.defaultTunnel
	}
	_, err := cloudflare.Attach(cloudflare.AttachOptions{
		Domain:     r.Hostname,
		LocalURL:   r.Origin,
		TunnelName: tunnel,
		ConfigDir:  h.configDir,
		Log:        h.log,
		Runner:     h.runner,
	})
	if err != nil {
		return "", err
	}
	return publicURL(r.Hostname), nil
}

func (h *hostBackend) RemoveRoute(ctx context.Context, tunnel, hostname string) error {
	_ = ctx
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if tunnel == "" {
		tunnel = h.defaultTunnel
	}
	return cloudflare.Detach(cloudflare.DetachOptions{
		Domain:     hostname,
		TunnelName: tunnel,
		ConfigDir:  h.configDir,
		Log:        h.log,
		Runner:     h.runner,
	})
}
