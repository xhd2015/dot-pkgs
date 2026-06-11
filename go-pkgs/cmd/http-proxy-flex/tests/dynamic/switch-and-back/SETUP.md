## Steps

1. Build the `http-proxy` binary
2. Start `http-proxy` pointing at `http://127.0.0.1:19998` with `--fallback-direct` and `--listen-port 19999`
3. Wait for initial "falling back to direct" log (upstream is dead at startup)
4. Start a TCP listener on port 19998 (simulate upstream becoming available)
5. Wait for "upstream proxy available, switching" log
6. Keep upstream available for 3 seconds (simulates "available for 5s"; test uses shorter duration)
7. Close the TCP listener (simulate upstream going down)
8. Wait for second "falling back to direct" log
9. Kill http-proxy and collect all output

```go
import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

func Run(t *testing.T, req *Request) (*Response, error) {
	binPath := getBinPath(t)

	cmd := exec.Command(binPath,
		"--upstream-proxy", "http://127.0.0.1:19998",
		"--fallback-direct",
		"--listen-port", "19999",
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

	// Step 3: wait for initial "falling back to direct"
	if !waitForPattern(func() string { return scNewOutput(sc) }, "falling back to direct", 10*time.Second) {
		return fail("timed out waiting for initial 'falling back to direct'")
	}
	scConsume(sc)

	// Step 4: start upstream listener
	upstream, err := net.Listen("tcp", "127.0.0.1:19998")
	if err != nil {
		return fail(fmt.Sprintf("start upstream listener: %v", err))
	}
	defer upstream.Close()

	// Step 5: wait for "upstream proxy available, switching"
	if !waitForPattern(func() string { return scNewOutput(sc) }, "upstream proxy available, switching", 10*time.Second) {
		upstream.Close()
		return fail("timed out waiting for 'upstream proxy available, switching'")
	}
	scConsume(sc)

	// Step 6: keep upstream available for 3 seconds
	time.Sleep(3 * time.Second)

	// Step 7: close upstream listener
	upstream.Close()

	// Step 8: wait for second "falling back to direct"
	if !waitForPattern(func() string { return scNewOutput(sc) }, "falling back to direct", 10*time.Second) {
		cmd.Process.Kill()
		cmd.Wait()
		return fail("timed out waiting for second 'falling back to direct'")
	}
	scConsume(sc)

	// Step 9: finalize
	time.Sleep(500 * time.Millisecond)
	cmd.Process.Kill()
	cmd.Wait()

	return &Response{
		Output:   scFullOutput(sc),
		ExitCode: 0,
	}, nil
}
```
