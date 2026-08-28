## Steps

1. Reserve random upstream and proxy ports
2. Start a TCP listener on the upstream port (upstream reachable at bootstrap)
3. Start `http-proxy` (default flex)
4. Wait for "using upstream proxy" and "listening on" logs
5. Close the upstream listener (upstream no longer listening)
6. Immediately send a CONNECT request — before the 1s health check can switch to direct
7. Kill http-proxy and capture all output

This reproduces the reported bug: proxy still routes CONNECT via upstream and logs
`connect to upstream proxy ... connection refused` instead of falling back to direct.

```go
import (
	"fmt"
	"net"
	"os/exec"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Bind :0 and keep the listener — do not reserveTCPPort()+rebind; that
	// close→listen gap races other parallel doctests for the same port.
	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("start upstream listener: %w", err)
	}
	upstreamPort := upstream.Addr().(*net.TCPAddr).Port
	proxyPort := reserveTCPPort(t)

	binPath := getBinPath(t, d)
	cmd := exec.Command(binPath,
		"--upstream-proxy", fmt.Sprintf("http://127.0.0.1:%d", upstreamPort),

		"--listen-port", fmt.Sprintf("%d", proxyPort),
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		upstream.Close()
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		upstream.Close()
		return err
	}
	defer cmd.Process.Kill()
	defer cmd.Wait()

	sc := newStreamCollector(stdout)

	if !waitForPattern(func() string { return scNewOutput(sc) }, "using upstream proxy", 10*time.Second) {
		upstream.Close()
		return fmt.Errorf("timed out waiting for 'using upstream proxy'\noutput:\n%s", scFullOutput(sc))
	}
	if !waitForPattern(func() string { return scNewOutput(sc) }, "listening on", 10*time.Second) {
		upstream.Close()
		return fmt.Errorf("timed out waiting for 'listening on'\noutput:\n%s", scFullOutput(sc))
	}
	scConsume(sc)

	upstream.Close()

	req.ConnectTarget = startLocalConnectTarget(t)
	req.ConnectResponse = connectThroughProxy(t, proxyPort, req.ConnectTarget)
	time.Sleep(500 * time.Millisecond)

	cmd.Process.Kill()
	cmd.Wait()

	req.CapturedOutput = scFullOutput(sc)
	return nil
}
```