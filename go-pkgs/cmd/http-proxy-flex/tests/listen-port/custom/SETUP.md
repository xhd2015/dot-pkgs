## Steps

- Run with `--listen-port 7829`, `--upstream-proxy` pointing to a dead port, and `--fallback-direct`
- The proxy starts, prints "listening on :7829", read for 2s, then kill

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Run(t *testing.T, req *Request) (*Response, error) {
	srcDir := filepath.Join(DOCTEST_ROOT, "..")
	binPath := filepath.Join(os.TempDir(), "http-proxy-test")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = srcDir
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build failed:\n%s", string(buildOut))
	}

	cmd := exec.Command(binPath,
		"--listen-port", "7829",
		"--upstream-proxy", "http://127.0.0.1:19999",
		"--fallback-direct",
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
