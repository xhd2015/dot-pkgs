# Scenario

**Feature**: default flex detects upstream becoming available after dead startup

```
# flex proxy starts with dead upstream, uses direct
http-proxy --upstream-proxy URL -> direct transport

# upstream later starts listening on URL
upstream proxy starts listening

# expected: flex proxy detects and switches to upstream
health monitor -> upstream proxy available, switching
```

## Steps

1. Reserve random upstream and proxy ports (nothing listening on upstream port)
2. Start `http-proxy` with default flex (no flags) pointing at the dead upstream port
3. Wait for initial "falling back to direct" and "listening on" logs
4. Start a TCP listener on the upstream port (upstream becomes available)
5. Wait for "upstream proxy available, switching" log
6. Kill http-proxy and collect all output

```go
import (
	"fmt"
	"os/exec"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return withAddrInUseRetry(func() error {
		binPath := getBinPath(t, d)

		upstreamPort, err := reserveClosedPort()
		if err != nil {
			return fmt.Errorf("reserve upstream port: %w", err)
		}
		proxyPort, err := reserveClosedPort()
		if err != nil {
			return fmt.Errorf("reserve proxy port: %w", err)
		}

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
			return fmt.Errorf("timed out waiting for initial 'falling back to direct'\noutput:\n%s", scFullOutput(sc))
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

		if !waitForPattern(getOutput, "upstream proxy available, switching", 10*time.Second) {
			cmd.Process.Kill()
			cmd.Wait()
			return fmt.Errorf("timed out waiting for 'upstream proxy available, switching'\noutput:\n%s", scFullOutput(sc))
		}
		scConsume(sc)

		time.Sleep(500 * time.Millisecond)
		cmd.Process.Kill()
		cmd.Wait()

		req.CapturedOutput = scFullOutput(sc)
		return nil
	})
}
```
