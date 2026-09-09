package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type helper struct {
	cfg config
}

func (h *helper) keyPath() string {
	return h.cfg.Dir + "/guest_ed25519"
}

func (h *helper) sshBaseArgs() []string {
	return []string{
		"ssh",
		"-i", h.keyPath(),
		"-p", strconv.Itoa(h.cfg.SSHPort),
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "IdentitiesOnly=yes",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5",
		"-o", "LogLevel=ERROR",
		fmt.Sprintf("%s@127.0.0.1", h.cfg.User),
	}
}

// shellQuote wraps s in single quotes for a remote login-shell command line.
// OpenSSH joins remote argv with spaces and re-parses via the user's shell -c,
// so multi-word scripts must be one shell-quoted token (same as printf %q).
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// sshRemote runs a single remote command string (already shell-safe).
func (h *helper) sshRemote(remoteCmd string) (string, error) {
	args := append(h.sshBaseArgs(), "--", remoteCmd)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// guestCmd runs sudo -n <args> in the guest.
func (h *helper) guestCmd(args ...string) (string, error) {
	parts := make([]string, 0, 2+len(args))
	parts = append(parts, "sudo", "-n")
	for _, a := range args {
		parts = append(parts, shellQuote(a))
	}
	return h.sshRemote(strings.Join(parts, " "))
}

// guestSh runs sudo -n bash -c <script> with the script shell-quoted so SSH
// login-shell reparse cannot split on spaces/semicolons.
func (h *helper) guestSh(script string) (string, error) {
	return h.sshRemote("sudo -n bash -c " + shellQuote(script))
}

func (h *helper) qemuAlive() bool {
	data, err := os.ReadFile(h.cfg.QemuPIDFile)
	if err != nil {
		return false
	}
	pid := strings.TrimSpace(string(data))
	if pid == "" {
		return false
	}
	n, err := strconv.Atoi(pid)
	if err != nil || n <= 0 {
		return false
	}
	cmd := exec.Command("kill", "-0", strconv.Itoa(n))
	return cmd.Run() == nil
}

func (h *helper) guestOK() bool {
	if _, err := os.Stat(h.keyPath()); err != nil {
		return false
	}
	_, err := h.guestCmd("true")
	return err == nil
}

func (h *helper) cfPID() string {
	out, err := h.guestCmd("pgrep", "-x", "cloudflared")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line != "" {
			return line
		}
	}
	return ""
}

func (h *helper) loginPID() string {
	out, err := h.guestSh(`ps -eo pid,args | awk '/[c]loudflared .*tunnel login/ {print $1; exit}'`)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.TrimSuffix(lines[len(lines)-1], "\r"))
}

func (h *helper) certOK() bool {
	_, err := h.guestCmd("test", "-s", h.cfg.Cert)
	return err == nil
}

func (h *helper) cfBinOK() bool {
	_, err := h.guestCmd("test", "-x", h.cfg.CFBin)
	return err == nil
}

func (h *helper) findTunnelUUID(name string) string {
	script := fmt.Sprintf("export TUNNEL_ORIGIN_CERT=%s; %s tunnel list 2>/dev/null", h.cfg.Cert, h.cfg.CFBin)
	list, _ := h.guestSh(script)
	for _, line := range strings.Split(list, "\n") {
		fields := strings.Fields(strings.TrimSuffix(line, "\r"))
		if len(fields) >= 2 && fields[1] == name {
			return fields[0]
		}
	}
	return ""
}

func (h *helper) extractURL() string {
	out, err := h.guestCmd("cat", h.cfg.CFURLFile)
	if err == nil {
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
			if line != "" {
				return line
			}
		}
	}
	script := fmt.Sprintf(`for f in %s %s; do
  [ -f "$f" ] || continue
  u=$(grep -oE 'https://[a-z0-9.-]+\.(trycloudflare\.com|xhd2015\.xyz)' "$f" 2>/dev/null | tail -1)
  [ -n "$u" ] && echo "$u" && exit 0
done
exit 0`, h.cfg.CFLog, h.cfg.OldLog)
	out, _ = h.guestSh(script)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line != "" {
			return line
		}
	}
	return ""
}

func (h *helper) extractLoginURL() string {
	script := fmt.Sprintf(`u=$(grep -oE 'https://dash\.cloudflare\.com[^[:space:]]+' %s 2>/dev/null | tail -1)
[ -n "$u" ] && echo "$u" && exit 0
grep -oE 'https://[^[:space:]]+' %s 2>/dev/null | grep -v trycloudflare | tail -1
exit 0`, h.cfg.CFLoginLog, h.cfg.CFLoginLog)
	out, _ := h.guestSh(script)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.TrimSuffix(lines[len(lines)-1], "\r"))
}

func (h *helper) printStatus(origin string) {
	if h.qemuAlive() {
		pid := ""
		if data, err := os.ReadFile(h.cfg.QemuPIDFile); err == nil {
			pid = strings.TrimSpace(string(data))
		}
		fmt.Println("qemu_alive=yes")
		fmt.Printf("qemu_pid=%s\n", pid)
	} else {
		fmt.Println("qemu_alive=no")
		fmt.Println("qemu_pid=")
	}
	if h.guestOK() {
		fmt.Println("guest_ssh=ok")
	} else {
		fmt.Println("guest_ssh=fail")
	}
	fmt.Printf("origin=%s\n", origin)
	if !h.qemuAlive() || !h.guestOK() {
		fmt.Println("cf_alive=no")
		fmt.Println("cf_pid=")
		fmt.Println("cf_bin=missing")
		fmt.Println("url=")
		fmt.Println("cert=missing")
		return
	}
	if h.cfBinOK() {
		fmt.Println("cf_bin=ok")
	} else {
		fmt.Println("cf_bin=missing")
	}
	pid := h.cfPID()
	if pid != "" {
		fmt.Println("cf_alive=yes")
		fmt.Printf("cf_pid=%s\n", pid)
	} else {
		fmt.Println("cf_alive=no")
		fmt.Println("cf_pid=")
	}
	fmt.Printf("url=%s\n", h.extractURL())
	if h.certOK() {
		fmt.Println("cert=ok")
	} else {
		fmt.Println("cert=missing")
	}
}

func (h *helper) requireGuest(origin string) error {
	if !h.qemuAlive() || !h.guestOK() {
		h.printStatus(origin)
		return fmt.Errorf("guest not ready")
	}
	return nil
}

func (h *helper) ensureCF() error {
	if h.cfBinOK() {
		fmt.Println("cf_bin=ok")
		return nil
	}
	dl := fmt.Sprintf("https://github.com/cloudflare/cloudflared/releases/download/%s/cloudflared-linux-amd64", h.cfg.CFVersion)
	script := fmt.Sprintf(`set -e
DL=%q
tmp=/tmp/cloudflared.$$
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "$tmp" "$DL"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "$tmp" "$DL"
else
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -y -o Acquire::Check-Valid-Until=false
  apt-get install -y --no-install-recommends curl ca-certificates
  curl -fsSL -o "$tmp" "$DL"
fi
chmod +x "$tmp"
mv "$tmp" %s
%s --version >/dev/null
`, dl, h.cfg.CFBin, h.cfg.CFBin)
	if _, err := h.guestSh(script); err != nil {
		return fmt.Errorf("ensure cloudflared: %w", err)
	}
	fmt.Println("cf_bin=ok")
	return nil
}

func (h *helper) killCF() {
	pid := h.cfPID()
	if pid == "" {
		return
	}
	_, _ = h.guestCmd("kill", pid)
	time.Sleep(time.Second)
	_, _ = h.guestCmd("kill", "-9", pid)
}

func (h *helper) pkillCF() {
	_, _ = h.guestSh("pkill -x cloudflared || true; sleep 1; pkill -9 -x cloudflared || true")
	time.Sleep(time.Second)
}
