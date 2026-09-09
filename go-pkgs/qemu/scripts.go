package qemu

import (
	"fmt"
	"regexp"
	"strconv"

	_ "embed"
)

//go:embed scripts/guest_start.sh
var guestStartScript string

//go:embed scripts/guest_cloudflared.sh
var guestCloudflaredScript string

// Dry-run placeholders (script bodies are not dumped).
const (
	GuestStartScriptPlaceholder       = "<guest_start.sh>"
	GuestCloudflaredScriptPlaceholder = "<guest_cloudflared.sh>"
)

// GuestStartScript returns the guest start script with Config values injected
// as concrete top-level assignments (so remote hosts without env still work).
func (c Config) GuestStartScript() string {
	c = c.normalized()
	s := guestStartScript
	s = rewriteAssign(s, "DIR", c.Dir)
	s = rewriteAssign(s, "USER", c.User)
	s = rewriteAssign(s, "PORT", strconv.Itoa(c.SSHPort))
	s = rewriteAssign(s, "MEM", strconv.Itoa(c.MemMB))
	s = rewriteAssign(s, "CPUS", strconv.Itoa(c.CPUs))
	s = rewriteAssign(s, "ACCEL", c.Accel)
	s = rewriteAssign(s, "BACKING_NAME", c.BackingName)
	s = rewriteAssign(s, "IMAGE_URL", c.BackingURL)
	s = rewriteAssign(s, "PKGVER", c.PkgVersion)
	return s
}

// GuestCloudflaredScript returns the legacy bash guest cloudflared script with
// Config values injected. Prefer Manager CF* methods (qemu-cf-helper) instead.
func (c Config) GuestCloudflaredScript() string {
	c = c.normalized()
	s := guestCloudflaredScript
	s = rewriteAssign(s, "DIR", c.Dir)
	s = rewriteAssign(s, "USER", c.User)
	s = rewriteAssign(s, "PORT", strconv.Itoa(c.SSHPort))
	s = rewriteAssign(s, "CF_BIN", c.CFBin)
	s = rewriteAssign(s, "CF_PID", c.CFPid)
	s = rewriteAssign(s, "CF_LOG", c.CFLog)
	s = rewriteAssign(s, "CERT", c.CFCert)
	s = rewriteAssign(s, "CF_VER", c.CFVersion)
	s = rewriteAssign(s, "DEFAULT_ORIGIN", c.CFOriginURL)
	return s
}

func (c Config) normalized() Config {
	d := WorkDevConfig()
	if c.Dir == "" {
		c.Dir = d.Dir
	}
	if c.User == "" {
		c.User = d.User
	}
	if c.SSHPort == 0 {
		c.SSHPort = d.SSHPort
	}
	if c.MemMB == 0 {
		c.MemMB = d.MemMB
	}
	if c.CPUs == 0 {
		c.CPUs = d.CPUs
	}
	if c.BackingName == "" {
		c.BackingName = d.BackingName
	}
	if c.BackingURL == "" {
		c.BackingURL = d.BackingURL
	}
	if c.PkgVersion == "" {
		c.PkgVersion = d.PkgVersion
	}
	if c.CFOriginURL == "" {
		c.CFOriginURL = d.CFOriginURL
	}
	if c.CFBin == "" {
		c.CFBin = d.CFBin
	}
	if c.CFCert == "" {
		c.CFCert = d.CFCert
	}
	if c.CFLog == "" {
		c.CFLog = d.CFLog
	}
	if c.CFPid == "" {
		c.CFPid = d.CFPid
	}
	if c.CFVersion == "" {
		c.CFVersion = d.CFVersion
	}
	if c.Accel == "" {
		c.Accel = d.Accel
	}
	return c
}

func rewriteAssign(script, key, value string) string {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `=.*$`)
	return re.ReplaceAllString(script, key+"="+value)
}

// GuestSSHArgs returns ssh argv to reach the guest on the qemu host loopback.
func GuestSSHArgs(cfg Config, tty bool) []string {
	cfg = cfg.normalized()
	args := []string{
		"ssh",
		"-i", cfg.Dir + "/guest_ed25519",
		"-p", strconv.Itoa(cfg.SSHPort),
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "IdentitiesOnly=yes",
		"-o", "LogLevel=ERROR",
	}
	if tty {
		args = append(args, "-tt")
	} else {
		args = append(args, "-o", "BatchMode=yes", "-o", "ConnectTimeout=5")
	}
	args = append(args, fmt.Sprintf("%s@127.0.0.1", cfg.User))
	return args
}
