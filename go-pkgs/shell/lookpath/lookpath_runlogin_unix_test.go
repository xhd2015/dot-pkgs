//go:build unix

package lookpath

import (
	"strings"
	"testing"
	"time"
)

// Live probe: defaultRunLogin must finish under a normal TTY session without
// leaving the test process Stopped (Setsid avoids SIGTTOU from bash -lic).
func TestDefaultRunLogin_BashCompletes(t *testing.T) {
	run := defaultRunLogin(8 * time.Second)
	out, err := run("bash", "command -v bash", minimalLoginEnv(""))
	if err != nil {
		t.Fatalf("defaultRunLogin bash: %v", err)
	}
	got := strings.TrimSpace(out)
	if got == "" {
		t.Fatal("empty stdout from command -v bash")
	}
	if !strings.Contains(got, "bash") {
		t.Fatalf("stdout %q does not look like a bash path", got)
	}
}
