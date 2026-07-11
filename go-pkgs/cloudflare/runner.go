package cloudflare

import (
	"fmt"
	"os/exec"
)

// DefaultRunner executes commands via os/exec and returns combined output.
type DefaultRunner struct{}

// Exec runs name with args and returns combined stdout+stderr.
func (DefaultRunner) Exec(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return out, err
}

// LookPathRunner is used by Status: Exec("cloudflared") with no args means look-path.
// DefaultRunner does not special-case zero args; Status uses exec.LookPath when Runner is nil,
// or Runner.Exec("cloudflared") when provided (tests inject a fake that understands this).

func runnerOrDefault(r CommandRunner) CommandRunner {
	if r != nil {
		return r
	}
	return DefaultRunner{}
}

func execCloudflared(r CommandRunner, args ...string) ([]byte, error) {
	return runnerOrDefault(r).Exec("cloudflared", args...)
}

// errOutput combines command output and error for soft-success checks.
func errOutput(out []byte, err error) string {
	s := string(out)
	if err != nil {
		if s != "" {
			return s + "\n" + err.Error()
		}
		return err.Error()
	}
	return s
}

func fmtCmdErr(op string, out []byte, err error) error {
	msg := string(out)
	if msg == "" && err != nil {
		msg = err.Error()
	}
	return fmt.Errorf("%s: %s", op, msg)
}
