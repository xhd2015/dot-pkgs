// Package backend is a unified cloudflared control plane: host-local or
// qemu-guest connector. Callers (e.g. ai-critic) select via Select.Qemu from
// qemu.json. When qemu is selected, cloudflared must not appear in host `ps`.
package backend

import (
	"context"
	"fmt"
	"io"

	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
	"github.com/xhd2015/dot-pkgs/go-pkgs/qemu"
)

// Kind identifies which connector implementation is in use.
type Kind string

const (
	KindHost Kind = "host"
	KindQemu Kind = "qemu"
)

// DefaultQemuTunnel is used when Select.DefaultTunnel is empty and Kind is qemu.
const DefaultQemuTunnel = "ai-critic-core"

// Status reports whether this backend can UpsertRoute.
type Status struct {
	Kind          Kind
	Available     bool
	Installed     bool
	Authenticated bool
	Detail        string
}

// Route is one public hostname mapped to an origin URL.
type Route struct {
	Tunnel   string // logical tunnel name; empty → backend default
	Hostname string // e.g. hello-world.xhd2015.xyz
	Origin   string // e.g. http://127.0.0.1:18080 or http://10.0.2.2:18080
}

// Backend hides whether cloudflared runs on the host or inside a qemu guest.
type Backend interface {
	Kind() Kind
	Status(ctx context.Context) (Status, error)
	EnsureReady(ctx context.Context) error
	UpsertRoute(ctx context.Context, r Route) (publicURL string, err error)
	RemoveRoute(ctx context.Context, tunnel, hostname string) error
	OriginForHostPort(port int) string
}

// Select chooses host vs qemu. App code reads qemu.json and sets Qemu.
type Select struct {
	Qemu bool

	HostConfigDir string
	HostRunner    cloudflare.CommandRunner
	Log           io.Writer

	QemuManager *qemu.Manager // nil → LocalHost + WorkDevConfig

	DefaultTunnel string
}

// New returns a host or qemu Backend per Select.Qemu.
func New(s Select) Backend {
	if s.Qemu {
		return newQemuBackend(s)
	}
	return newHostBackend(s)
}

// FromQemuEnabled is a convenience for ai-critic: qemu.json enabled flag.
func FromQemuEnabled(enabled bool) Backend {
	return New(Select{Qemu: enabled})
}

func validateRoute(r Route) error {
	if r.Hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if r.Origin == "" {
		return fmt.Errorf("origin is required")
	}
	return nil
}

func publicURL(hostname string) string {
	return "https://" + hostname
}
