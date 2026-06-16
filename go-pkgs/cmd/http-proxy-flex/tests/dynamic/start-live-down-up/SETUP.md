## Steps

1. Start a TCP listener on port 19996 to simulate an available upstream proxy
2. Build and start `http-proxy` pointing at `http://127.0.0.1:19996` with `--fallback-direct` and `--listen-port 19997`
3. Wait for initial "using upstream proxy" log (upstream is live at startup)
4. Close the TCP listener (simulate upstream going down)
5. Wait for "falling back to direct" log
6. Start a new TCP listener on port 19996 (simulate upstream becoming available again)
7. Wait for "upstream proxy available, switching" log
8. Kill http-proxy and collect all output

```go
import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	binPath := getBinPath(t)

	// Step 1: start upstream listener before the proxy
	upstream, err := net.Listen("tcp", "127.0.0.1:19996")
	if err != nil {
		return fmt.Errorf("start upstream listener: %v", err)
	}
	defer upstream.Close()

	// Step 2: start http-proxy
	cmd := exec.Command(binPath,
		"--upstream-proxy", "http://127.0.0.1:19996",
		"--fallback-direct",
		"--listen-port", "19997",
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

	// Step 3: wait for initial "using upstream proxy"
	if !waitForPattern(getOutput, "using upstream proxy", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("timed out waiting for 'using upstream proxy'\noutput:\n%s", scFullOutput(sc))
	}
	scConsume(sc)

	// Step 4: close upstream listener (simulate going down)
	upstream.Close()

	// Step 5: wait for "falling back to direct"
	if !waitForPattern(getOutput, "falling back to direct", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("timed out waiting for 'falling back to direct'\noutput:\n%s", scFullOutput(sc))
	}
	scConsume(sc)

	// Step 6: start new upstream listener on same port
	upstream2, err := net.Listen("tcp", "127.0.0.1:19996")
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("start upstream listener again: %v", err)
	}
	defer upstream2.Close()

	// Step 7: wait for "upstream proxy available, switching"
	if !waitForPattern(getOutput, "upstream proxy available, switching", 10*time.Second) {
		upstream2.Close()
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("timed out waiting for 'upstream proxy available, switching'\noutput:\n%s", scFullOutput(sc))
	}
	scConsume(sc)

	// Step 8: finalize
	time.Sleep(500 * time.Millisecond)
	cmd.Process.Kill()
	cmd.Wait()

	req.CapturedOutput = scFullOutput(sc)
	return nil
}
```
