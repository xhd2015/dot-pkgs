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

func Run(t *testing.T, req *Request) (*Response, error) {
	binPath := getBinPath(t)

	// Step 1: start upstream listener before the proxy
	upstream, err := net.Listen("tcp", "127.0.0.1:19996")
	if err != nil {
		return nil, fmt.Errorf("start upstream listener: %v", err)
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
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	sc := newStreamCollector(stdout)

	fail := func(msg string) (*Response, error) {
		cmd.Process.Kill()
		cmd.Wait()
		return nil, fmt.Errorf("%s\noutput so far:\n%s", msg, scFullOutput(sc))
	}

	// Step 3: wait for initial "using upstream proxy"
	if !waitForPattern(func() string { return scNewOutput(sc) }, "using upstream proxy", 10*time.Second) {
		return fail("timed out waiting for 'using upstream proxy'")
	}
	scConsume(sc)

	// Step 4: close upstream listener (simulate going down)
	upstream.Close()

	// Step 5: wait for "falling back to direct"
	if !waitForPattern(func() string { return scNewOutput(sc) }, "falling back to direct", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		return fail("timed out waiting for 'falling back to direct'")
	}
	scConsume(sc)

	// Step 6: start new upstream listener on same port
	upstream2, err := net.Listen("tcp", "127.0.0.1:19996")
	if err != nil {
		return fail(fmt.Sprintf("start upstream listener again: %v", err))
	}
	defer upstream2.Close()

	// Step 7: wait for "upstream proxy available, switching"
	if !waitForPattern(func() string { return scNewOutput(sc) }, "upstream proxy available, switching", 10*time.Second) {
		upstream2.Close()
		return fail("timed out waiting for 'upstream proxy available, switching'")
	}
	scConsume(sc)

	// Step 8: finalize
	time.Sleep(500 * time.Millisecond)
	cmd.Process.Kill()
	cmd.Wait()

	return &Response{
		Output:   scFullOutput(sc),
		ExitCode: 0,
	}, nil
}
```
