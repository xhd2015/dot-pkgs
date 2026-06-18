# Scenario

Dynamic upstream health: dead at startup → comes up → goes down again.

## Steps

1. Build the `http-proxy` binary
2. Reserve random upstream and listen ports (no listener on upstream port yet)
3. Start `http-proxy` pointing at the reserved upstream port with `--fallback-direct`
4. Wait for initial "falling back to direct" and "listening on" logs (upstream dead at startup)
5. Start a TCP listener on the upstream port (simulate upstream becoming available)
6. Wait for "upstream proxy available, switching" log
7. Keep upstream available for 3 seconds
8. Close the TCP listener (simulate upstream going down)
9. Wait for second "falling back to direct" log
10. Kill http-proxy and collect all output

```go
import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	binPath := getBinPath(t)

	// Reserve upstream port but keep it closed at startup.
	reserved, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("reserve upstream port: %w", err)
	}
	upstreamPort := reserved.Addr().(*net.TCPAddr).Port
	reserved.Close()

	proxyLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
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

	upstream, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", upstreamPort))
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("start upstream listener: %v", err)
	}
	defer upstream.Close()

	if !waitForPattern(getOutput, "upstream proxy available, switching", 10*time.Second) {
		upstream.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("timed out waiting for 'upstream proxy available, switching'\noutput:\n%s", scFullOutput(sc))
	}
	scConsume(sc)

	time.Sleep(3 * time.Second)

	upstream.Close()

	if !waitForPattern(getOutput, "falling back to direct", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("timed out waiting for second 'falling back to direct'\noutput:\n%s", scFullOutput(sc))
	}
	scConsume(sc)

	time.Sleep(500 * time.Millisecond)
	cmd.Process.Kill()
	cmd.Wait()

	req.CapturedOutput = scFullOutput(sc)
	return nil
}