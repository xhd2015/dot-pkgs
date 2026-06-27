## Steps

1. Reserve random upstream and proxy ports (nothing listening on upstream port)
2. Start `http-proxy` (default flex) pointing at the dead upstream port
3. Wait for "falling back to direct" and "listening on" logs
4. Send a CONNECT request through the proxy
5. Kill http-proxy and capture all output

```go
import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	upstreamPort := reserveTCPPort(t)
	proxyPort := reserveTCPPort(t)

	binPath := getBinPath(t)
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
	defer cmd.Process.Kill()
	defer cmd.Wait()

	sc := newStreamCollector(stdout)

	if !waitForPattern(func() string { return scNewOutput(sc) }, "listening on", 10*time.Second) {
		return fmt.Errorf("timed out waiting for 'listening on'\noutput:\n%s", scFullOutput(sc))
	}
	if !waitForPattern(func() string { return scNewOutput(sc) }, "falling back to direct", 10*time.Second) {
		return fmt.Errorf("timed out waiting for 'falling back to direct'\noutput:\n%s", scFullOutput(sc))
	}
	scConsume(sc)

	req.ConnectTarget = startLocalConnectTarget(t)
	req.ConnectResponse = connectThroughProxy(t, proxyPort, req.ConnectTarget)
	time.Sleep(500 * time.Millisecond)

	cmd.Process.Kill()
	cmd.Wait()

	req.CapturedOutput = scFullOutput(sc)
	return nil
}
```