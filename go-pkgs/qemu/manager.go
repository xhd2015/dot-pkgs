package qemu

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Manager runs guest qemu / cloudflared scripts on a Host.
type Manager struct {
	Cfg    Config
	Host   Host
	DryRun bool
	// optional writers for dry-run lines; if nil discard
	Stdout io.Writer
	Stderr io.Writer
}

func (m *Manager) host() Host {
	if m.Host != nil {
		return m.Host
	}
	return LocalHost{}
}

func (m *Manager) stdout() io.Writer {
	if m.Stdout != nil {
		return m.Stdout
	}
	return io.Discard
}

func (m *Manager) cfOrigin(origin string) string {
	if origin != "" {
		return origin
	}
	c := m.Cfg.normalized()
	return c.CFOriginURL
}

func (m *Manager) dryLine(placeholder string, action string, extra ...string) string {
	parts := append([]string{action}, extra...)
	return fmt.Sprintf("sh -c %s -- %s", placeholder, strings.Join(parts, " "))
}

func (m *Manager) dryLineHelper(action string, extra ...string) string {
	parts := append([]string{HelperPlaceholder}, HelperFlags(m.Cfg)...)
	parts = append(parts, action)
	parts = append(parts, extra...)
	return strings.Join(parts, " ")
}

func (m *Manager) runGuest(action string, extra ...string) error {
	if m.DryRun {
		fmt.Fprintf(m.stdout(), "[dry-run] would %s\n", m.dryLine(GuestStartScriptPlaceholder, action, extra...))
		return nil
	}
	// Scripts are bash (use [[, etc.); Debian /bin/sh is dash and exits 2.
	args := append([]string{"-c", m.Cfg.GuestStartScript(), "--", action}, extra...)
	return m.host().Run("bash", args...)
}

func (m *Manager) runGuestOut(action string, extra ...string) (string, error) {
	if m.DryRun {
		fmt.Fprintf(m.stdout(), "[dry-run] would %s\n", m.dryLine(GuestStartScriptPlaceholder, action, extra...))
		return "", nil
	}
	args := append([]string{"-c", m.Cfg.GuestStartScript(), "--", action}, extra...)
	return m.host().Output("bash", args...)
}

func (m *Manager) runCF(action string, extra ...string) error {
	if m.DryRun {
		fmt.Fprintf(m.stdout(), "[dry-run] would %s\n", m.dryLineHelper(action, extra...))
		return nil
	}
	bin, err := m.EnsureHelper()
	if err != nil {
		return err
	}
	args := append(HelperFlags(m.Cfg), action)
	args = append(args, extra...)
	return m.host().Run(bin, args...)
}

func (m *Manager) runCFOut(action string, extra ...string) (string, error) {
	if m.DryRun {
		fmt.Fprintf(m.stdout(), "[dry-run] would %s\n", m.dryLineHelper(action, extra...))
		return "", nil
	}
	bin, err := m.EnsureHelper()
	if err != nil {
		return "", err
	}
	args := append(HelperFlags(m.Cfg), action)
	args = append(args, extra...)
	return m.host().Output(bin, args...)
}

// Status probes qemu pid, guest ssh, and disk.
func (m *Manager) Status() (map[string]string, error) {
	out, err := m.runGuestOut("status")
	return ParseKV(out), err
}

// Start brings the guest up idempotently.
func (m *Manager) Start() (map[string]string, error) {
	out, err := m.runGuestOut("start")
	return ParseKV(out), err
}

// Stop stops qemu; overlay is kept.
func (m *Manager) Stop() error {
	return m.runGuest("stop")
}

// Restart stops then starts the guest.
func (m *Manager) Restart() (map[string]string, error) {
	if err := m.Stop(); err != nil {
		return nil, err
	}
	return m.Start()
}

// Purge stops qemu and deletes the overlay; deep also deletes the backing image.
func (m *Manager) Purge(deep bool) error {
	extra := []string{}
	if deep {
		extra = append(extra, "--deep")
	}
	return m.runGuest("purge", extra...)
}

// Logs returns the guest serial log tail.
func (m *Manager) Logs(lines int) (string, error) {
	if lines <= 0 {
		lines = 80
	}
	return m.runGuestOut("logs", strconv.Itoa(lines))
}

// EnsureRunning returns status if qemu is alive, otherwise Start.
func (m *Manager) EnsureRunning() (map[string]string, error) {
	kv, err := m.Status()
	if m.DryRun {
		return kv, err
	}
	if err == nil && kv["qemu_alive"] == "yes" {
		return kv, nil
	}
	return m.Start()
}

// CFStatus probes guest cloudflared.
func (m *Manager) CFStatus(origin string) (map[string]string, error) {
	out, err := m.runCFOut("status", m.cfOrigin(origin))
	return ParseKV(out), err
}

// CFStart starts a quick tunnel to origin.
func (m *Manager) CFStart(origin string) (map[string]string, error) {
	out, err := m.runCFOut("start", m.cfOrigin(origin))
	return ParseKV(out), err
}

// CFStop kills guest cloudflared.
func (m *Manager) CFStop() error {
	return m.runCF("stop")
}

// CFRestart stops then starts a quick tunnel.
func (m *Manager) CFRestart(origin string) (map[string]string, error) {
	if err := m.CFStop(); err != nil {
		return nil, err
	}
	return m.CFStart(origin)
}

// CFPurge stops cloudflared and clears logs/pid/url; deep also deletes cert.pem.
func (m *Manager) CFPurge(deep bool) error {
	extra := []string{}
	if deep {
		extra = append(extra, "--deep")
	}
	return m.runCF("purge", extra...)
}

// CFLogs returns the guest cloudflared log tail.
func (m *Manager) CFLogs(lines int) (string, error) {
	if lines <= 0 {
		lines = 80
	}
	return m.runCFOut("logs", strconv.Itoa(lines))
}

// CFLogin runs cloudflared tunnel login in the guest and waits for cert.pem.
func (m *Manager) CFLogin() (map[string]string, error) {
	out, err := m.runCFOut("login")
	return ParseKV(out), err
}

// CFList runs cloudflared tunnel list in the guest.
func (m *Manager) CFList() (string, error) {
	return m.runCFOut("list")
}

// CFNamedStart runs named-start NAME ORIGIN host1 host2...
// Prefer CFNamedApply when hostnames need distinct origin URLs.
func (m *Manager) CFNamedStart(tunnelName, origin string, hostnames []string) (map[string]string, error) {
	origin = m.cfOrigin(origin)
	extra := append([]string{tunnelName, origin}, hostnames...)
	out, err := m.runCFOut("named-start", extra...)
	return ParseKV(out), err
}

// GuestOriginForPort returns the SLIRP host-gateway origin for a process on the qemu host.
func GuestOriginForPort(port int) string {
	return fmt.Sprintf("http://10.0.2.2:%d", port)
}

// CFNamedApply ensures a named tunnel in the guest with per-host origins,
// routes DNS, and restarts guest cloudflared.
// routes maps hostname → origin URL.
func (m *Manager) CFNamedApply(tunnelName string, hostnames []string, routes map[string]string, _ string) error {
	if tunnelName == "" {
		return fmt.Errorf("tunnel name is required")
	}
	if len(routes) == 0 {
		return fmt.Errorf("at least one route is required")
	}
	pairs := make([]string, 0, len(routes))
	for _, h := range hostnames {
		origin := routes[h]
		if origin == "" {
			continue
		}
		pairs = append(pairs, h+"="+origin)
	}
	if len(pairs) == 0 {
		for h, origin := range routes {
			pairs = append(pairs, h+"="+origin)
		}
		sort.Strings(pairs)
	}
	extra := append([]string{tunnelName}, pairs...)
	out, err := m.runCFOut("named-apply", extra...)
	// Script may still exit non-zero after a successful apply (e.g. set -e
	// race around status probes); treat applied=yes as success.
	if strings.Contains(out, "applied=yes") || strings.Contains(out, "named=yes") {
		return nil
	}
	return err
}

// GuestReadFile returns the contents of path inside the guest (via sudo cat).
func (m *Manager) GuestReadFile(path string) (string, error) {
	cfg := m.Cfg.normalized()
	args := GuestSSHArgs(cfg, false)
	args = append(args, "--", "sudo", "-n", "cat", path)
	return m.host().Output(args[0], args[1:]...)
}
