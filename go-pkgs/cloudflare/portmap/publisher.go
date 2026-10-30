package portmap

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
)

// ManagedPublisher uses a dedicated caller-owned named tunnel. One wildcard
// ingress forwards only to Manager's host-allowlisted relay. Exact DNS entries
// are created on demand. Adding/expiring sibling ports does NOT restart the
// connector, so existing WebSockets and streams remain connected.
type ManagedPublisher struct {
	TunnelName string
	ConfigDir  string
	DNS        DNSController
	Log        io.Writer
	// Step optionally receives one progress line per publish stage, e.g.
	// "verifying Cloudflare domain example.com". Publish calls it while
	// holding its lock: the callback must not re-enter the publisher, and a
	// slow body delays provisioning.
	Step           func(format string, args ...any)
	mu             sync.Mutex
	hosts          map[string]map[string]bool
	ensureTunnelFn func(context.Context) (string, error)
	attachFn       func(context.Context, string, string) error
	detachFn       func(context.Context, string) error
}

func (p *ManagedPublisher) step(format string, args ...any) {
	if p.Step != nil {
		p.Step(format, args...)
	}
}

// DNSController permits fake API tests as well as the origin-cert DNS client.
type DNSController interface {
	VerifyDomain(context.Context, string) error
	EnsureTunnelCNAME(context.Context, string, string) error
}

type contextRunner struct{ ctx context.Context }

func (r contextRunner) Exec(name string, args ...string) ([]byte, error) {
	return exec.CommandContext(r.ctx, name, args...).CombinedOutput()
}

// Attach must start a real managed connector, so its Runner is intentionally nil.

func (p *ManagedPublisher) configDir() (string, error) {
	if p.ConfigDir != "" {
		return p.ConfigDir, nil
	}
	return cloudflare.DefaultConfigDir()
}
func (p *ManagedPublisher) dns() DNSController {
	if p.DNS != nil {
		return p.DNS
	}
	dir, _ := p.configDir()
	return &cloudflare.OriginCertDNSClient{CertPath: filepath.Join(dir, "cert.pem")}
}

// Reset removes ingress left by a previous process. Call before accepting
// mapping requests. Never use a tunnel shared with another application.
func (p *ManagedPublisher) Reset(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if strings.TrimSpace(p.TunnelName) == "" {
		return fmt.Errorf("a dedicated tunnel name is required")
	}
	dir, err := p.configDir()
	if err != nil {
		return err
	}
	managed, err := cloudflare.ManagedTunnelDir(dir, p.TunnelName)
	if err != nil {
		return err
	}
	state, err := cloudflare.LoadTunnelState(managed)
	if err != nil {
		return err
	}
	for host := range state.Hosts {
		if err := cloudflare.Detach(cloudflare.DetachOptions{Domain: host, TunnelName: p.TunnelName, ConfigDir: dir, Log: p.Log}); err != nil {
			return err
		}
	}
	p.hosts = map[string]map[string]bool{}
	return ctx.Err()
}

func (p *ManagedPublisher) Publish(ctx context.Context, host, origin string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.TunnelName == "" {
		return fmt.Errorf("a dedicated tunnel name is required")
	}
	// Each published hostname is one DNS label below the configured zone.
	dot := strings.IndexByte(host, '.')
	if dot < 1 {
		return fmt.Errorf("hostname must have a base domain")
	}
	domain := host[dot+1:]
	wildcard := "*." + domain
	p.step("verifying Cloudflare domain %s…", domain)
	if err := p.dns().VerifyDomain(ctx, domain); err != nil {
		return err
	}
	p.step("ensuring dedicated tunnel %s…", p.TunnelName)
	tunnelID, err := p.ensureTunnel(ctx)
	if err != nil {
		return err
	}
	p.step("publishing DNS %s…", host)
	if err := p.dns().EnsureTunnelCNAME(ctx, host, tunnelID); err != nil {
		return err
	}
	if p.hosts == nil {
		p.hosts = map[string]map[string]bool{}
	}
	if len(p.hosts[wildcard]) == 0 {
		p.step("starting tunnel connector…")
		err = p.attach(ctx, wildcard, origin)
		if err != nil {
			return err
		}
		p.hosts[wildcard] = map[string]bool{}
	}
	p.hosts[wildcard][host] = true
	return nil
}

func (p *ManagedPublisher) ensureTunnel(ctx context.Context) (string, error) {
	if p.ensureTunnelFn != nil {
		return p.ensureTunnelFn(ctx)
	}
	id, _, err := cloudflare.EnsureTunnelExists(contextRunner{ctx: ctx}, p.TunnelName)
	return id, err
}
func (p *ManagedPublisher) attach(ctx context.Context, wildcard, origin string) error {
	if p.attachFn != nil {
		return p.attachFn(ctx, wildcard, origin)
	}
	_, err := cloudflare.Attach(cloudflare.AttachOptions{Domain: wildcard, LocalURL: origin, TunnelName: p.TunnelName, ConfigDir: p.ConfigDir, Log: p.Log, SkipDNS: true})
	return err
}
func (p *ManagedPublisher) detach(ctx context.Context, wildcard string) error {
	if p.detachFn != nil {
		return p.detachFn(ctx, wildcard)
	}
	return cloudflare.Detach(cloudflare.DetachOptions{Domain: wildcard, TunnelName: p.TunnelName, ConfigDir: p.ConfigDir, Log: p.Log})
}

// Unpublish never deletes DNS or someone else's tunnel. The relay refuses the
// expired hostname immediately. The last route also detaches the wildcard and
// stops the dedicated connector; harmless owned CNAMEs remain for later reuse.
func (p *ManagedPublisher) Unpublish(ctx context.Context, host string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for wildcard, hosts := range p.hosts {
		if !hosts[host] {
			continue
		}
		if len(hosts) == 1 {
			if err := p.detach(ctx, wildcard); err != nil {
				return err
			}
			delete(p.hosts, wildcard)
		} else {
			delete(hosts, host)
		}
	}
	return ctx.Err()
}
