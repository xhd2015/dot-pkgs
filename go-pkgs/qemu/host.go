package qemu

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Host executes commands on the machine that hosts qemu (not inside the guest).
type Host interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
}

// LineHandler receives one stdout line with the trailing newline removed.
type LineHandler func(line string)

// StreamingHost delivers stdout lines before the process exits.
// Hosts that do not implement it are buffered by Manager until exit.
type StreamingHost interface {
	OutputLines(ctx context.Context, name string, args []string, onLine LineHandler) (string, error)
}

// LocalHost runs commands via os/exec on the local machine.
type LocalHost struct{}

// Run executes name with args and returns its exit error.
func (LocalHost) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// Output runs name with args and returns combined stdout/stderr.
func (LocalHost) Output(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// OutputLines runs name and calls onLine for each stdout line as soon as it
// is written. Cancelling ctx kills the process group so a hung ssh does not
// outlive the command. Stderr is not mixed into the line stream.
func (LocalHost) OutputLines(ctx context.Context, name string, args []string, onLine LineHandler) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = 2 * time.Second
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return "", err
	}
	var buf strings.Builder
	sc := bufio.NewScanner(stdout)
	for sc.Scan() {
		line := sc.Text()
		buf.WriteString(line)
		buf.WriteByte('\n')
		if onLine != nil {
			onLine(line)
		}
	}
	scanErr := sc.Err()
	waitErr := cmd.Wait()
	out := buf.String()
	if scanErr != nil {
		return out, scanErr
	}
	if waitErr != nil {
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		if stderr.Len() > 0 {
			return out, fmt.Errorf("%w: %s", waitErr, strings.TrimSpace(stderr.String()))
		}
		return out, waitErr
	}
	return out, nil
}
