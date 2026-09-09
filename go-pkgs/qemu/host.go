package qemu

import "os/exec"

// Host executes commands on the machine that hosts qemu (not inside the guest).
type Host interface {
	Run(name string, args ...string) error
	Output(name string, args ...string) (string, error)
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
