## Steps

1. Start a local HTTP test server as the request target
2. Start a fake TCP listener as the upstream proxy
3. Build and start http-proxy pointing at the fake upstream
4. Wait for "using upstream proxy" and "listening on" logs
5. Make an HTTP GET request through the proxy to the test server
6. Wait briefly for the log line to appear
7. Kill http-proxy, return all captured output

```go
import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func Run(t *testing.T, req *Request) (*Response, error) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer targetServer.Close()

	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start fake upstream: %w", err)
	}
	defer upstream.Close()
	upstreamAddr := upstream.Addr().String()

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("find proxy port: %w", err)
	}
	proxyPort := proxyListener.Addr().(*net.TCPAddr).Port
	proxyListener.Close()

	srcDir := filepath.Join(DOCTEST_ROOT, "..")
	binPath := filepath.Join(os.TempDir(), "http-proxy-test")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = srcDir
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed:\n%s", string(buildOut))
	}

	cmd := exec.Command(binPath,
		"--upstream-proxy", "http://"+upstreamAddr,
		"--listen-port", fmt.Sprintf("%d", proxyPort),
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer cmd.Process.Kill()
	defer cmd.Wait()

	var mu sync.Mutex
	var outputBuf strings.Builder
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			mu.Lock()
			outputBuf.WriteString(scanner.Text() + "\n")
			mu.Unlock()
		}
	}()

	getOutput := func() string {
		mu.Lock()
		defer mu.Unlock()
		return outputBuf.String()
	}

	if !waitForPattern(getOutput, "listening on", 10*time.Second) {
		return nil, fmt.Errorf("timed out waiting for 'listening on'\noutput:\n%s", getOutput())
	}

	proxyURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", proxyPort))
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 3 * time.Second,
	}
	client.Get(targetServer.URL + "/test")

	time.Sleep(1 * time.Second)

	cmd.Process.Kill()
	cmd.Wait()

	return &Response{
		Output:   getOutput(),
		ExitCode: 0,
	}, nil
}

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
```
