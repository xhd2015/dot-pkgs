## Preconditions

- The `http-proxy` Go module exists at `../` relative to the test root
- `go build` can compile the binary

## Steps

1. Build the `http-proxy` binary once (cached across tests)
2. Execute the binary with the arguments provided in `req.Args`
3. Capture combined stdout+stderr output and exit code

## Context

- The binary is a forward HTTP proxy that listens on a local port
- It supports `--listen-port`, `--upstream-proxy`, `--fallback-direct`, and `--help`

```go
import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Request struct {
	Args           []string
	CapturedOutput string
}

type Response struct {
	Output   string
	ExitCode int
}

// --- binary build (sync.Once, cached across all tests) ---

var (
	buildOnce sync.Once
	cachedBin string
	buildErr  error
)

func getBinPath(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		srcDir := filepath.Join(DOCTEST_ROOT, "..")
		cachedBin = filepath.Join(os.TempDir(), "http-proxy-test")
		buildCmd := exec.Command("go", "build", "-o", cachedBin, ".")
		buildCmd.Dir = srcDir
		out, err := buildCmd.CombinedOutput()
		if err != nil {
			buildErr = fmt.Errorf("build failed:\n%s", string(out))
		}
	})
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	return cachedBin
}

// --- start/stop helper for short-lived process tests ---

func startAndCapture(t *testing.T, binPath string, args ...string) string {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 4096)
	time.Sleep(2 * time.Second)
	n, _ := output.Read(buf)
	cmd.Process.Kill()
	cmd.Wait()
	return string(buf[:n])
}

// --- streaming collector for long-running process tests ---

type streamCollector struct {
	mu       sync.Mutex
	buf      strings.Builder
	consumed int
}

func newStreamCollector(r io.Reader) *streamCollector {
	s := &streamCollector{}
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			s.mu.Lock()
			s.buf.WriteString(scanner.Text() + "\n")
			s.mu.Unlock()
		}
	}()
	return s
}

func scNewOutput(sc *streamCollector) string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	out := sc.buf.String()
	if sc.consumed < len(out) {
		return out[sc.consumed:]
	}
	return ""
}

func scConsume(sc *streamCollector) {
	sc.mu.Lock()
	sc.consumed = sc.buf.Len()
	sc.mu.Unlock()
}

func scFullOutput(sc *streamCollector) string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.buf.String()
}

// --- waitForPattern helper ---

func waitForPattern(getOutput func() string, pattern string, timeout time.Duration) bool {
	deadline := time.After(timeout)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			return false
		case <-ticker.C:
			if strings.Contains(getOutput(), pattern) {
				return true
			}
		}
	}
}

// default Run — handles both short-lived commands (e.g. --help) and long-running server processes
func Run(t *testing.T, req *Request) (*Response, error) {
	if req.CapturedOutput != "" {
		return &Response{
			Output:   req.CapturedOutput,
			ExitCode: 0,
		}, nil
	}

	binPath := getBinPath(t)
	cmd := exec.Command(binPath, req.Args...)

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stdoutBuf

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		// command exited naturally (e.g. --help)
	case <-time.After(4 * time.Second):
		// long-running server — kill and capture output
		cmd.Process.Kill()
		<-done
	}

	return &Response{
		Output:   stdoutBuf.String(),
		ExitCode: 0,
	}, nil
}
```
