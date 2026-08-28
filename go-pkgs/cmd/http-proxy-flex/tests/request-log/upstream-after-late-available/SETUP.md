# Scenario

**Bug**: HTTP requests stay on direct after upstream becomes available post-startup

```
# proxy starts with dead upstream (default flex)
http-proxy -> direct transport

# upstream later available; client sends GET
upstream starts -> GET /test -> expected via upstream proxy
```

## Steps

1. Start a local HTTP test server as the request target
2. Reserve upstream and proxy ports (upstream dead at startup)
3. Start `http-proxy` with default flex (no flags)
4. Wait for "falling back to direct" and "listening on"
5. Start a TCP listener on the upstream port
6. Wait 2 seconds for health monitor to detect upstream (if working)
7. Make an HTTP GET through the proxy
8. Kill http-proxy and capture all output

```go
import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return withAddrInUseRetry(func() error {
		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}))
		defer targetServer.Close()

		upstreamPort := reserveTCPPort(t)
		proxyPort := reserveTCPPort(t)

		binPath := getBinPath(t, d)
		cmd := exec.Command(binPath,
			"--upstream-proxy", fmt.Sprintf("http://127.0.0.1:%d", upstreamPort),
			"--listen-port", fmt.Sprintf("%d", proxyPort),
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		cmd.Stderr = cmd.Stdout

		if err := cmd.Start(); err != nil {
			return err
		}

		sc := newStreamCollector(stdout)
		getOutput := func() string { return scNewOutput(sc) }

		if !waitForPattern(getOutput, "falling back to direct", 10*time.Second) {
			cmd.Process.Kill()
			cmd.Wait()
			return fmt.Errorf("timed out waiting for 'falling back to direct'\noutput:\n%s", scFullOutput(sc))
		}
		if !waitForPattern(getOutput, "listening on", 10*time.Second) {
			cmd.Process.Kill()
			cmd.Wait()
			return fmt.Errorf("timed out waiting for 'listening on'\noutput:\n%s", scFullOutput(sc))
		}
		scConsume(sc)

		upstream, err := listenExactPort(upstreamPort)
		if err != nil {
			cmd.Process.Kill()
			cmd.Wait()
			return fmt.Errorf("start upstream listener: %w", err)
		}
		defer upstream.Close()

		time.Sleep(2 * time.Second)
		scConsume(sc)

		proxyURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", proxyPort))
		client := &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
			Timeout:   3 * time.Second,
		}
		client.Get(targetServer.URL + "/test")
		time.Sleep(1 * time.Second)

		cmd.Process.Kill()
		cmd.Wait()

		req.CapturedOutput = scFullOutput(sc)
		return nil
	})
}
```