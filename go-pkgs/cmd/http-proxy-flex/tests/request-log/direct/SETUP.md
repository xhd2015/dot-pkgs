## Steps

1. Start a local HTTP test server as the request target
2. Build and start http-proxy with a dead upstream and `--fallback-direct`
3. Wait for "falling back to direct" and "listening on" logs
4. Make an HTTP GET request through the proxy to the test server
5. Wait briefly for the log line to appear
6. Kill http-proxy, return all captured output

```go
import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"time"
)

func Run(t *testing.T, req *Request) (*Response, error) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer targetServer.Close()

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("find proxy port: %w", err)
	}
	proxyPort := proxyListener.Addr().(*net.TCPAddr).Port
	proxyListener.Close()

	binPath := getBinPath(t)

	cmd := exec.Command(binPath,
		"--upstream-proxy", "http://127.0.0.1:19999",
		"--fallback-direct",
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

	sc := newStreamCollector(stdout)

	if !waitForPattern(func() string { return scNewOutput(sc) }, "listening on", 10*time.Second) {
		return nil, fmt.Errorf("timed out waiting for 'listening on'\noutput:\n%s", scFullOutput(sc))
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
		Output:   scFullOutput(sc),
		ExitCode: 0,
	}, nil
}
```
