## Steps

1. Build the `http-proxy` binary
2. Start `http-proxy` pointing at `http://127.0.0.1:19998` with `--fallback-direct` and `--listen-port 19999`
3. Wait for initial "falling back to direct" log (upstream is dead at startup)
4. Start a TCP listener on port 19998 (simulate upstream becoming available)
5. Wait for "upstream proxy available, switching" log
6. Keep upstream available for 3 seconds (simulates "available for 5s"; test uses shorter duration)
7. Close the TCP listener (simulate upstream going down)
8. Wait for second "falling back to direct" log
9. Kill http-proxy and collect all output

```go
import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func Run(t *testing.T, req *Request) (*Response, error) {
	srcDir := filepath.Join(DOCTEST_ROOT, "..")
	binPath := filepath.Join(os.TempDir(), "http-proxy-test")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = srcDir
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed:\n%s", string(buildOut))
	}

	cmd := exec.Command(binPath,
		"--upstream-proxy", "http://127.0.0.1:19998",
		"--fallback-direct",
		"--listen-port", "19999",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var mu sync.Mutex
	var outputBuf strings.Builder
	consumed := 0

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			mu.Lock()
			outputBuf.WriteString(scanner.Text() + "\n")
			mu.Unlock()
		}
	}()

	newOutput := func() string {
		mu.Lock()
		defer mu.Unlock()
		s := outputBuf.String()
		if consumed < len(s) {
			return s[consumed:]
		}
		return ""
	}

	consume := func() {
		mu.Lock()
		consumed = outputBuf.Len()
		mu.Unlock()
	}

	fullOutput := func() string {
		mu.Lock()
		defer mu.Unlock()
		return outputBuf.String()
	}

	fail := func(msg string) (*Response, error) {
		cmd.Process.Kill()
		cmd.Wait()
		return nil, fmt.Errorf("%s\noutput so far:\n%s", msg, fullOutput())
	}

	// Step 3: wait for initial "falling back to direct"
	if !waitForPattern(newOutput, "falling back to direct", 10*time.Second) {
		return fail("timed out waiting for initial 'falling back to direct'")
	}
	consume()

	// Step 4: start upstream listener
	upstream, err := net.Listen("tcp", "127.0.0.1:19998")
	if err != nil {
		return fail(fmt.Sprintf("start upstream listener: %v", err))
	}
	defer upstream.Close()

	// Step 5: wait for "upstream proxy available, switching"
	if !waitForPattern(newOutput, "upstream proxy available, switching", 10*time.Second) {
		upstream.Close()
		return fail("timed out waiting for 'upstream proxy available, switching'")
	}
	consume()

	// Step 6: keep upstream available for 3 seconds
	time.Sleep(3 * time.Second)

	// Step 7: close upstream listener
	upstream.Close()

	// Step 8: wait for second "falling back to direct"
	if !waitForPattern(newOutput, "falling back to direct", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		return fail("timed out waiting for second 'falling back to direct'")
	}
	consume()

	// Step 9: finalize
	time.Sleep(500 * time.Millisecond)
	cmd.Process.Kill()
	cmd.Wait()

	return &Response{
		Output:   fullOutput(),
		ExitCode: 0,
	}, nil
}

func waitForPattern(getNewOutput func() string, pattern string, timeout time.Duration) bool {
	deadline := time.After(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			return false
		case <-ticker.C:
			if strings.Contains(getNewOutput(), pattern) {
				return true
			}
		}
	}
}
```
