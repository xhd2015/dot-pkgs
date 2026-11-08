// Package elevate runs shell scripts with root privileges on macOS.
//
// Three strategies are supported, in increasing order of intrusion:
//
//   - ModeSudo: `sudo -- /bin/sh -c <script>`, reusing an existing sudo ticket
//     and prompting in the controlling TTY (/dev/tty) when one is available.
//   - ModeOsaScript: the native macOS administrator dialog. The script travels
//     base64-encoded, so no AppleScript quoting is involved.
//
// The password never passes through this package: sudo reads it from the
// terminal and osascript owns its own dialog.
package elevate

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Mode selects how root privileges are obtained.
type Mode string

const (
	// ModeSudo prompts in the controlling terminal via sudo.
	ModeSudo Mode = "sudo"
	// ModeOsaScript shows the native macOS administrator password dialog.
	ModeOsaScript Mode = "osascript"
)

// DefaultPrompt is the password prompt prefix used when Config.Prompt is empty.
const DefaultPrompt = "clear-tmp"

// Config is one privileged script invocation.
type Config struct {
	Purpose string    // short label, e.g. "dirs-cleaner purge"
	Script  string    // shell script, executed with Shell
	Mode    Mode      // ModeSudo (default) or ModeOsaScript
	Shell   string    // default /bin/sh
	Prompt  string    // prompt prefix, default DefaultPrompt
	Stdin   *os.File  // optional; defaults to /dev/tty then os.Stdin
	Stdout  io.Writer // default os.Stdout
	Stderr  io.Writer // default os.Stderr
}

func (c Config) prompt() string {
	if c.Prompt != "" {
		return c.Prompt
	}
	return DefaultPrompt
}

func (c Config) shell() string {
	if c.Shell != "" {
		return c.Shell
	}
	return "/bin/sh"
}

// Run executes c.Script as root and returns its standard output.
func Run(c Config) ([]byte, error) {
	if c.Mode == ModeOsaScript {
		return osaScript(c)
	}
	return sudo(c)
}

// RunShell executes script as the current user and returns its stdout.
func RunShell(script string) ([]byte, error) {
	c := exec.Command("/bin/sh", "-c", script)
	var stderr bytes.Buffer
	c.Stderr = &stderr
	out, err := c.Output()
	if err != nil {
		return out, fmt.Errorf("script failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func sudo(c Config) ([]byte, error) {
	prompt := "Password for " + c.prompt() + " " + c.Purpose + ": "
	cmd := exec.Command("sudo", "-p", prompt, "--", c.shell(), "-c", c.Script)
	if c.Stdin != nil {
		cmd.Stdin = c.Stdin
	} else if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		cmd.Stdin = tty
		defer tty.Close()
	} else {
		cmd.Stdin = os.Stdin
	}
	if c.Stderr != nil {
		cmd.Stderr = c.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("sudo %s: %w", c.Purpose, err)
	}
	return out, nil
}

// osaScript shows the macOS administrator dialog. The script travels
// base64-encoded so no AppleScript quoting is involved; the password never
// passes through the caller.
func osaScript(c Config) ([]byte, error) {
	b64 := base64.StdEncoding.EncodeToString([]byte(c.Script))
	as := fmt.Sprintf(`do shell script "echo %s | /usr/bin/base64 -D | /bin/sh" with prompt "%s: %s" with administrator privileges without altering line endings`,
		b64, c.prompt(), c.Purpose)
	cmd := exec.Command("osascript", "-e", as)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if strings.Contains(msg, "-128") {
			return nil, fmt.Errorf("%s: administrator authorization cancelled", c.Purpose)
		}
		return nil, fmt.Errorf("%s: osascript: %v: %s", c.Purpose, err, msg)
	}
	return out, nil
}

// IsRoot reports whether the process runs as root.
func IsRoot() bool { return os.Geteuid() == 0 }

// IsTerminal reports whether an interactive terminal is available. /dev/tty is
// preferred so sudo can prompt even when stdout is piped.
func IsTerminal() bool {
	if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
		_ = f.Close()
		return true
	}
	return isCharDevice(os.Stdin)
}

// CanWriteDir reports whether the current process may create files in dir.
// A missing dir is treated as writable: there is nothing to remove.
func CanWriteDir(dir string) bool {
	st, err := os.Stat(dir)
	if err != nil {
		return os.IsNotExist(err)
	}
	if !st.IsDir() {
		return false
	}
	f, err := os.CreateTemp(dir, ".elevate-probe-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}

func isCharDevice(f *os.File) bool {
	if f == nil {
		return false
	}
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return (st.Mode() & os.ModeCharDevice) != 0
}
