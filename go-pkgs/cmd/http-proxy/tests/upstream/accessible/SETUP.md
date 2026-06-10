## Steps

- Start a TCP listener on a random port to simulate an available upstream proxy
- Run `http-proxy` pointing at that listener
- Capture initial log output, then kill

```go
import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Run(t *testing.T, req *Request) (*Response, error) {
	// Start a fake upstream listener
	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start upstream listener: %w", err)
	}
	defer upstream.Close()
	upstreamAddr := upstream.Addr().String()

	srcDir := filepath.Join(DOCTEST_ROOT, "..")
	binPath := filepath.Join(os.TempDir(), "http-proxy-test")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = srcDir
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed:\n%s", string(buildOut))
	}

	cmd := exec.Command(binPath,
		"--upstream-proxy", "http://"+upstreamAddr,
	)

	output, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	buf := make([]byte, 4096)
	time.Sleep(2 * time.Second)
	n, _ := output.Read(buf)

	cmd.Process.Kill()
	cmd.Wait()

	return &Response{
		Output:   string(buf[:n]),
		ExitCode: 0,
	}, nil
}
```
