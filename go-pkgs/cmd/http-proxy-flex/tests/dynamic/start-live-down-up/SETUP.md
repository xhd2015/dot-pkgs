# Scenario

Dynamic upstream health: live at startup → goes down → comes back up.

## Steps

1. Start a TCP listener on a random port to simulate an available upstream proxy
2. Build and start `http-proxy` pointing at that port with `--fallback-direct` on a random listen port
3. Wait for "using upstream proxy" and "listening on" logs (upstream live at startup)
4. Close the TCP listener (simulate upstream going down)
5. Wait for "falling back to direct" log
6. Start a new TCP listener on the same port (simulate upstream becoming available again)
7. Wait for "upstream proxy available, switching" log
8. Kill http-proxy and collect all output

```go
import (
	"fmt"
	"net"
	"os/exec"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	binPath := getBinPath(t)

	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("start upstream listener: %v", err)
	}
	upstreamPort := upstream.Addr().(*net.TCPAddr).Port

	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		upstream.Close()
		return fmt.Errorf("reserve proxy port: %w", err)
	}
	proxyPort := proxyLn.Addr().(*net.TCPAddr).Port
	proxyLn.Close()

	cmd := exec.Command(binPath,
		"--upstream-proxy", fmt.Sprintf("http://127.0.0.1:%d", upstreamPort),
		"--fallback-direct",
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

	sc := newStreamCollector(stdout)
	getOutput := func() string { return scNewOutput(sc) }

	if !waitForPattern(getOutput, "using upstream proxy", 10*time.Second) {
		upstream.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("timed out waiting for 'using upstream proxy'\noutput:\n%s", scFullOutput(sc))
	}
	if !waitForPattern(getOutput, "listening on", 10*time.Second) {
		upstream.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("timed out waiting for 'listening on'\noutput:\n%s", scFullOutput(sc))
	}
	scConsume(sc)

	upstream.Close()

	if !waitForPattern(getOutput, "falling back to direct", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("timed out waiting for 'falling back to direct'\noutput:\n%s", scFullOutput(sc))
	}
	scConsume(sc)

	upstream2, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", upstreamPort))
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("start upstream listener again: %v", err)
	}
	defer upstream2.Close()

	if !waitForPattern(getOutput, "upstream proxy available, switching", 10*time.Second) {
		upstream2.Close()
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
}
```